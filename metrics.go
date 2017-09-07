package main

import (
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rancher/container-crontab/events"
)

var (
	activeJobGauge *prometheus.GaugeVec
)

func initMetrics() {
	hostname, _ := os.Hostname()
	activeJobGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "rancher_container_crontab_jobs",
			Help:        "Number of container crontab job entries",
			ConstLabels: prometheus.Labels{"hostname": hostname},
		}, []string{"state"})
	prometheus.MustRegister(activeJobGauge)
}

func collectMetrics(handler *events.DockerHandler) {
	for {
		handler.GetJobStats(activeJobGauge)
		time.Sleep(5 * time.Second)
	}
}

func MetricsServer(handler *events.DockerHandler) {
	initMetrics()

	go collectMetrics(handler)

	http.Handle("/metrics", promhttp.Handler())
	logrus.Fatal(http.ListenAndServe(":9191", nil))
}
