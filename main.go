package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {

	c, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	go tailDockerEvents(c)

	sC := make(chan os.Signal, 1)
	signal.Notify(sC, os.Interrupt)

	log.Println("watching...")

	<-sC

	log.Println("exiting")
}

func tailDockerEvents(client *client.Client) {

	f := filters.NewArgs(

		filters.Arg("event", "start"),
		filters.Arg("event", "destroy"),
		filters.Arg("event", "stop"),
		filters.Arg("event", "die"),
		filters.Arg("event", "kill"),
		filters.Arg("event", "oom"))

	o := &types.EventsOptions{Filters: f}

	m, errC := client.Events(context.Background(), *o)

	for {
		select {
		case i := <-m:

			switch i.Action {
			case "start":
				if _, ok := i.Actor.Attributes["ecs.healthcheck.enabled"]; !ok {
					break
				}

				if _, ok := i.Actor.Attributes["com.amazonaws.ecs.task-arn"]; !ok {
					log.Println("task has healtcheck enabled but does not have a task-arn attribute - skipping")
					break
				}

				container, err := client.ContainerInspect(context.Background(), i.ID)

				if err != nil {
					log.Println("error getting container spec - skipping")
					log.Println(err)
					break
				}

				if container.NetworkSettings == nil {
					log.Println("cannot get network settings - skipping")
					break
				}

				h := &healthCheck{
					id:          i.ID,
					ip:          container.NetworkSettings.IPAddress,
					path:        i.Actor.Attributes["ecs.healthcheck.path"],
					interval:    i.Actor.Attributes["ecs.healthcheck.interval"],
					failedCount: i.Actor.Attributes["ecs.healthcheck.failed"],
					taskArn:     i.Actor.Attributes["com.amazonaws.ecs.task-arn"],
				}

				if h.ip == "" {
					log.Println("cannot get container ip - skipping")
					break
				}

				if h.path == "" {
					h.path = "/health-check"
				}

				if h.interval == "" {
					h.interval = "30"
				}

				if h.failedCount == "" {
					h.failedCount = "2"
				}

				log.Println("scheduling healthcheck for", h)

				go func(p *healthCheck) {
					t := time.NewTicker(5 * time.Second)
					p.ticker = t
					p.done = make(chan bool, 1)

					(*healthChecks)[p.id] = p
				F:
					for {
						select {
						case <-t.C:
							log.Println("ping", p.ip, p.id)
						case <-p.done:
							log.Println("done")
							break F
						}
					}

				}(h)

			case "destroy", "stop", "die", "kill", "oom":
				if _, ok := i.Actor.Attributes["ecs.healthcheck.enabled"]; !ok {
					break
				}

				h, found := (*healthChecks)[i.ID]

				if !found {
					log.Println("healthcheck not found - skipping", i.ID)
					break
				}

				h.ticker.Stop()
				h.done <- true

				delete(*healthChecks, i.ID)

				log.Println("descheduled health check")
			}

		case e := <-errC:
			log.Fatal(e)
		}
	}
}

var healthChecks = &map[string]*healthCheck{}

type healthCheck struct {
	id          string
	ip          string
	path        string
	interval    string
	failedCount string
	taskArn     string
	ticker      *time.Ticker
	done        chan bool
}
