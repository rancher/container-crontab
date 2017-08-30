package cron

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"gopkg.in/robfig/cron.v2"
)

type cronJob interface {
	Deactivate()
}

// Crontab is the struct that holds the cron runner
type Crontab struct {
	cronRunner *cron.Cron
	jobs       map[string]cron.EntryID
}

// NewCrontab creates the crontab
func NewCrontab() (*Crontab, error) {
	logrus.Infof("Starting Cron")
	crontab := &Crontab{
		cronRunner: cron.New(),
		jobs:       map[string]cron.EntryID{},
	}

	crontab.cronRunner.Start()

	return crontab, nil
}

// GetEntries lists the cron entries
func (ct *Crontab) GetEntries() []cron.Entry {
	entries := ct.cronRunner.Entries()
	return entries
}

// AddJob Adds a docker job to the crontab
func (ct *Crontab) AddJob(id string, labels map[string]string, jobType string) error {
	var job cron.Job

	if ct.jobs[id] != 0 {
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
	ct.jobs[id] = jobID
	logrus.Infof("Added: %s, with schedule: %s", id, schedule)
	return nil
}

// RemoveJob remove a docker job from the cron queue
func (ct *Crontab) RemoveJob(id string) {
	if _, ok := ct.jobs[id]; ok {
		ct.cronRunner.Remove(ct.jobs[id])
		delete(ct.jobs, id)
		logrus.Infof("Removed: %s", id)
	}
}
