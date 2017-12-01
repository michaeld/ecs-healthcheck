package main

import (
	"context"
	"log"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func TestDockerEvents(t *testing.T) {
	c, err := client.NewEnvClient()

	if err != nil {
		panic(err)
	}

	o := &types.EventsOptions{}
	ctx := context.Background()
	m, _ := c.Events(ctx, *o)

	message := <-m

	log.Println(message)
}

/*
   "Labels": {
       "com.amazonaws.ecs.cluster": "sandbox1",
       "com.amazonaws.ecs.container-name": "subscription-service",
       "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-east-1:089964245684:task/bff49df6-14ca-4351-a1b3-9b0f894d7aa1",
       "com.amazonaws.ecs.task-definition-family": "subscription-service-sandbox",
       "com.amazonaws.ecs.task-definition-version": "27"
   }
*/
