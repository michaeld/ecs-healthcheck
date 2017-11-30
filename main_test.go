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
