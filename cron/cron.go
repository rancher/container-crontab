package cron

import (
	"github.com/Sirupsen/logrus"
	"github.com/robfig/cron"
)

type cronJob interface {
	Deactivate()
}

// Crontab is the struct that holds the cron runner
type Crontab struct {
	cronRunner *cron.Cron
	jobs       map[string]cronJob
}

// NewCrontab creates the crontab
func NewCrontab() (*Crontab, error) {
	logrus.Infof("Starting Cron")
	crontab := &Crontab{
		cronRunner: cron.New(),
		jobs:       map[string]cronJob{},
	}

	crontab.cronRunner.Start()

	return crontab, nil
}

// GetEntries lists the cron entries
func (ct *Crontab) GetEntries() []*cron.Entry {
	entries := ct.cronRunner.Entries()
	return entries
}

// AddJob Adds a docker job to the crontab
func (ct *Crontab) AddJob(id string, labels map[string]string, jobType string) {
	var schedule string
	var job cron.Job

	if ct.jobs[id] != nil {
		logrus.Debugf("Job %s, already exists", id)
		return
	}

	if sched, ok := labels["cron.schedule"]; ok {
		schedule = sched
	}

	switch jobType {
	case "docker":
		dj := NewDockerJob(id, labels)
		ct.jobs[id] = dj
		job = dj
	default:
		logrus.Warnf("Unknown job type: %s", jobType)
	}
	err := ct.cronRunner.AddJob(schedule, job)
	if err != nil {
		logrus.Errorf("error adding: %s. Got: %s", id, err)
	} else {
		logrus.Infof("Added: %s, with schedule: %s", id, schedule)
	}
}

// RemoveJob remove a docker job from the cron queue
func (ct *Crontab) RemoveJob(id string) {
	if job, ok := ct.jobs[id]; ok {
		job.Deactivate()
		logrus.Infof("Removed: %s", id)
	}

}
