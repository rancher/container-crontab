package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/container-crontab/events"
	"github.com/urfave/cli"
)

// VERSION of the application
var (
	MetadataURL = "http://169.254.169.250/2016-07-29"
	VERSION     = "v0.0.0-dev"
)

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "container-crontab"
	app.Version = VERSION
	app.Usage = "container-crontab"
	app.Action = start
	app.Before = beforeApp
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "debug,d",
		},
		cli.BoolFlag{
			Name:  "rancher-mode,r",
			Usage: "Allow Rancher ",
		},
		cli.StringFlag{
			Name:  "metadata-url",
			Value: MetadataURL,
			Usage: "Provide full URL of Metadata",
		},
		cli.BoolFlag{
			Name: "metrics",
		},
	}

	app.Run(os.Args)
}

func start(c *cli.Context) error {
	handler, err := events.NewDockerHandler(&events.DockerHandlerOpts{
		RancherMode: c.GlobalBool("rancher-mode"),
		MetadataURL: c.GlobalString("metadata-url"),
	})
	if err != nil {
		return err
	}

	router, err := events.NewEventRouter()
	if err != nil {
		return err
	}

	if c.GlobalBool("metrics") {
		go MetricsServer(handler)
	}

	events.StartRouter(router, handler)

	return nil
}
