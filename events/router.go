package events

import (
	"context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Router Interface
type Router interface {
	Listen(context.Context) (<-chan events.Message, <-chan error)
}

//DockerEventRouter is the Docker event handler implementation
type DockerEventRouter struct {
	DockerClient *client.Client
	Handler      Handler
}

// NewEventRouter returns the a Docker event handler
func NewEventRouter() (Router, error) {
	dClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return DockerEventRouter{
		DockerClient: dClient,
	}, nil
}

// StartRouter calls the listener function and takes the interface for testing
func StartRouter(router Router, handler Handler) {
loop:
	for {
		ctx, cancelFunc := context.WithCancel(context.Background())
		eventStream, errChan := router.Listen(ctx)
		select {
		case event := <-eventStream:
			handler.Handle(&event)
		case err := <-errChan:
			logrus.Error(err)
			cancelFunc()
			continue loop
		}
	}
}

// Listen implements the Router interface
func (de DockerEventRouter) Listen(ctx context.Context) (<-chan events.Message, <-chan error) {
	filterArgs := filters.NewArgs()
	// Adds the cron job
	filterArgs.Add("event", "start")
	filterArgs.Add("event", "create")

	filterArgs.Add("event", "stop")
	filterArgs.Add("event", "die")

	// removes from the cron queue
	filterArgs.Add("event", "destroy")

	eventOptions := types.EventsOptions{
		Filters: filterArgs,
	}

	return de.DockerClient.Events(ctx, eventOptions)
}
