package events

import (
	"context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rancher/container-crontab/cron"
)

// Router Interface
type Router interface {
	Listen() (<-chan events.Message, <-chan error)
}

// Handler handles messages
type Handler interface {
	Handle(Message)
}

// Message is a message from an event stream
type Message *events.Message

// DockerHandler handles docker messages
type DockerHandler struct {
	crontab *cron.Crontab
}

// NewDockerHandler returns a docker handler with crontab
func NewDockerHandler() *DockerHandler {
	crontab, _ := cron.NewCrontab()

	dClient, err := client.NewClient("unix:///var/run/docker.sock", "1.22", nil, map[string]string{})
	if err != nil {
		logrus.Fatal(err)
		return nil
	}

	containers, err := dClient.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		logrus.Fatal(err)
		return nil
	}

	// Scan containers
	logrus.Infof("Scanning for container cron entries")
	for _, container := range containers {
		if _, ok := container.Labels["cron.schedule"]; ok {
			err = crontab.AddJob(container.ID, container.Labels, "docker")
			if err != nil {
				logrus.Errorf("error adding: %s. Got: %s", container.ID, err)
			}
		}
	}

	return &DockerHandler{
		crontab: crontab,
	}
}

// Handle implements handler interface
func (dh DockerHandler) Handle(msg Message) {
	// Adding a cron.schedule label flags the container for deeper inspection
	// With this service
	if _, ok := msg.Actor.Attributes["cron.schedule"]; ok {
		if msg.Action == "start" || msg.Action == "create" {
			logrus.Debugf("Processing %s event for container: %s", msg.Action, msg.ID)
			dh.crontab.AddJob(msg.ID, msg.Actor.Attributes, "docker")
		}

		if msg.Action == "destroy" {
			logrus.Debugf("Processing destroy event for container: %s", msg.ID)
			dh.crontab.RemoveJob(msg.ID)
		}
	}

	dh.crontab.GetEntries()
}

//DockerEventRouter is the Docker event handler implementation
type DockerEventRouter struct {
	DockerClient *client.Client
}

// NewEventRouter returns the a Docker event handler
func NewEventRouter() (Router, error) {
	dClient, err := client.NewClient("unix:///var/run/docker.sock", "1.22", nil, map[string]string{})
	if err != nil {
		return nil, err
	}
	return DockerEventRouter{
		DockerClient: dClient,
	}, nil
}

// StartRouter calls the listener function and takes the interface for testing
func StartRouter(router Router) {
	handler := NewDockerHandler()
	for {
		eventStream, errChan := router.Listen()
		select {
		case event := <-eventStream:
			handler.Handle(&event)
		case err := <-errChan:
			logrus.Error(err)
		}
	}
}

// Listen implements the Router interface
func (de DockerEventRouter) Listen() (<-chan events.Message, <-chan error) {
	filterArgs := filters.NewArgs()
	// Adds the cron job
	filterArgs.Add("event", "start")
	filterArgs.Add("event", "create")

	// Not sure that we need these...unknown usecases
	// filterArgs.Add("event", "stop")
	// filterArgs.Add("event", "die")

	// removes from the cron queue
	filterArgs.Add("event", "destroy")

	eventOptions := types.EventsOptions{
		Filters: filterArgs,
	}

	return de.DockerClient.Events(context.Background(), eventOptions)
}
