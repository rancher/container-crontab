package cron

import (
	"strings"
	"time"
)

func getDuration(i int) time.Duration {
	return time.Duration(i) * time.Second
}

func getRancherStackNameFromLabels(labels map[string]string) string {
	if stackName, ok := labels["io.rancher.stack.name"]; ok {
		return stackName
	}

	return ""
}

func getRancherServiceNameFromLabels(labels map[string]string) string {
	if stackServiceName, ok := labels["io.rancher.project_service.name"]; ok {
		serviceSplit := strings.SplitN(stackServiceName, "/", 2)
		if len(serviceSplit) == 2 {
			return serviceSplit[1]
		}
	}
	return ""
}
