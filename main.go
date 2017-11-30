package main

import (
	"context"
	"log"
	"os"
	"os/signal"

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
		filters.Arg("event", "create"),
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
			log.Println(i)
		case e := <-errC:
			log.Fatal(e)
		}
	}
}
