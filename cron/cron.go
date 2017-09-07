package cron

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
	"gopkg.in/robfig/cron.v2"
)

type cronJob interface {
	Deactivate()
}

// Crontab is the struct that holds the cron runner
type Crontab struct {
	cronRunner *cron.Cron
	jobs       map[string]*JobEntry
	mdClient   metadata.Client
	rancher    bool
}

type JobEntry struct {
	CronID cron.EntryID
	Job    *DockerJob
}

// NewCrontab creates the crontab
func NewCrontab() (*Crontab, error) {
	logrus.Infof("Starting Cron")
	crontab := &Crontab{
		cronRunner: cron.New(),
		jobs:       map[string]*JobEntry{},
	}

	crontab.cronRunner.Start()

	return crontab, nil
}

func NewRancherTypeCrontab(metadataURL string) (*Crontab, error) {
	crontab, err := NewCrontab()
	if err != nil {
		return crontab, nil
	}

	crontab.mdClient, err = metadata.NewClientAndWait(metadataURL)
	if err != nil {
		return crontab, nil
	}

	crontab.rancher = true

	go crontab.watchRancherMetadata()

	return crontab, nil
}

// GetEntries lists the cron entries
func (ct *Crontab) GetEntries() []cron.Entry {
	entries := ct.cronRunner.Entries()
	return entries
}

// AddJob Adds a docker job to the crontab
func (ct *Crontab) AddJob(id string, labels map[string]string, jobType string) error {
	var job *DockerJob

	if _, ok := ct.jobs[id]; ok {
		logrus.Debugf("Igoring Event: %d with job id: %d", id, ct.jobs[id])
		return nil
	}

	schedule, ok := labels["cron.schedule"]
	if !ok {
		return fmt.Errorf("No cron schedule found for container: %s", id)
	}

	switch jobType {
	case "docker":
		job = NewDockerJob(id, labels)
	default:
		logrus.Warnf("Unknown job type: %s", jobType)
	}

	jobID, err := ct.cronRunner.AddJob(schedule, job)
	if err != nil {
		logrus.Errorf("error adding: %s. Got: %s", id, err)
		return err
	}

	ct.jobs[id] = &JobEntry{
		CronID: jobID,
		Job:    job,
	}

	ct.setJobState(ct.jobs[id])

	logrus.Infof("Added: %s, with schedule: %s", id, schedule)
	return nil
}

// RemoveJob remove a docker job from the cron queue
func (ct *Crontab) RemoveJob(id string) {
	if jobEntry, ok := ct.jobs[id]; ok {
		ct.cronRunner.Remove(jobEntry.CronID)
		delete(ct.jobs, id)
		logrus.Infof("Removed: %s", id)
	}
}

func (ct *Crontab) DeactivateJob(id string, labels map[string]string) error {
	if !ct.rancher {
		return nil
	}

	if jobEntry, ok := ct.jobs[id]; ok {
		ct.setJobState(jobEntry)
	}

	return nil
}

func (ct *Crontab) checkRancherMetadataServiceState(stackName, serviceName string) (string, error) {
	stack, err := ct.mdClient.GetStackByName(stackName)
	if err != nil {
		return "", err
	}

	for _, service := range stack.Services {
		if service.Name == serviceName {
			return service.State, nil
		}
	}

	return "", nil
}

func (ct *Crontab) watchRancherMetadata() {
	for {
		logrus.Debug("Scanning Rancher Metadata")
		for _, job := range ct.jobs {
			ct.setJobState(job)
		}
		time.Sleep(getDuration(5))
	}
}

func (ct *Crontab) setJobState(job *JobEntry) {
	if !ct.rancher {
		return
	}

	stackName := getRancherStackNameFromLabels(job.Job.Labels)
	serviceName := getRancherServiceNameFromLabels(job.Job.Labels)
	if stackName != "" && serviceName != "" {
		state, err := ct.checkRancherMetadataServiceState(stackName, serviceName)
		if err != nil {
			logrus.Error(err)
		}

		// if the job is inactive...activate
		if state == "active" && !job.Job.Active {
			job.Job.Activate()
		}

		// if the job is active... Deactivate
		if state != "active" && job.Job.Active {
			job.Job.Deactivate()
		}
	}
}

func (ct *Crontab) GetNumberOfActiveJobs() float64 {
	var i float64
	for _, job := range ct.jobs {
		if job.Job.Active {
			i++
		}
	}
	return i
}

func (ct *Crontab) GetNumberOfInactiveJobs() float64 {
	var i float64
	for _, job := range ct.jobs {
		if !job.Job.Active {
			i++
		}
	}
	return i
}
