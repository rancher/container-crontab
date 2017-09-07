package cron

import (
	"context"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerJob implements the cron job interface
type DockerJob struct {
	ID             string
	Action         string
	Schedule       string
	Leader         bool
	Labels         map[string]string
	Active         bool
	lastError      error
	restartTimeout time.Duration
}

// Err returns last error message
func (dj *DockerJob) Err() error {
	return dj.lastError
}

// Run Implements the job interface from cron package
func (dj *DockerJob) Run() {
	defer dj.resetErr()

	if dj.Active {
		logrus.Debugf("Executing: %s on %s", dj.Action, dj.ID)
		switch dj.Action {
		case "start":
			dj.start()
		case "restart":
			dj.restart()
		case "stop":
			dj.stop()
		default:
			logrus.Errorf("Unsupported action: %s for container id: %s", dj.Action, dj.ID)
		}
	}

	if dj.Err() != nil {
		logrus.Error(dj.Err())
	}
}

func (dj *DockerJob) resetErr() {
	if dj.lastError != nil {
		logrus.Debugf("Reseting error on %s", dj.ID)
	}
	dj.lastError = nil
}

func (dj *DockerJob) start() {
	var client *client.Client
	client, dj.lastError = getDockerClient()
	defer client.Close()

	if dj.Err() == nil {
		dj.lastError = client.ContainerStart(context.Background(), dj.ID, types.ContainerStartOptions{})
	}
}

func (dj *DockerJob) restart() {
	var client *client.Client
	client, dj.lastError = getDockerClient()
	defer client.Close()

	if dj.Err() == nil {
		dj.lastError = client.ContainerRestart(context.Background(), dj.ID, &dj.restartTimeout)
	}
}

func (dj *DockerJob) stop() {
	var client *client.Client
	client, dj.lastError = getDockerClient()
	defer client.Close()

	if dj.Err() == nil {
		dj.lastError = client.ContainerStop(context.Background(), dj.ID, &dj.restartTimeout)
	}
}

func getDockerClient() (*client.Client, error) {
	return client.NewEnvClient()
}

// NewDockerJob creates a DockerJob and sets defaults
func NewDockerJob(id string, labels map[string]string) *DockerJob {
	dj := &DockerJob{
		ID:             id,
		Schedule:       labels["cron.schedule"],
		Labels:         labels,
		Action:         "start",
		Leader:         false,
		Active:         true,
		lastError:      nil,
		restartTimeout: getDuration(10),
	}

	if value, ok := labels["cron.action"]; ok {
		dj.Action = value
	}

	if _, ok := labels["cron.leader"]; ok {
		dj.Leader = true
	}

	if TO, ok := labels["cron.restart_timeout"]; ok {
		i, err := strconv.Atoi(TO)
		if err != nil {
			logrus.Error("Error converting cron.restart_timeout to int, sticking with default of 10seconds")
			logrus.Error(err)
			i = 10
		}
		dj.restartTimeout = getDuration(i)
	}

	return dj
}

// Deactivate Sets the Actve attribute to false. This will skip running
func (dj *DockerJob) Deactivate() {
	logrus.Debugf("Deactivating: %s", dj.ID)
	dj.Active = false
}

func (dj *DockerJob) Activate() {
	logrus.Debugf("Activating: %s", dj.ID)
	dj.Active = true
}
