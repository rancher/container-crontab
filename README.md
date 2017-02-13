container-crontab
========

A microservice that will perform actions on a Docker container based on cron schedule.

## Building

`make`

## Running

`./bin/container-crontab`

## Usage

Once `container-crontab` is up and running it watches Docker socket events for `create, start and destroy` events.
If a container is found to have the label `cron.schedule` then it will be added to the crontab based on the schedule.

Cron scheduling rules follow: [Expression Format](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format)

## Override labels that can be applied

To override the default start action on the container, set the label `cron.action` equal to ``stop` or `restart`.

To override the default 10 second restart/stop timeout set the label `cron.restart_timeout` to the number of
seconds you would like. For instance for 20 seconds: `cron.restart_timeout=20`.

## Examples
```
# Restart every minute
> docker run -d --label=cron.schedule="0 * * * * ?" ubuntu:16.04 date
```

## License
Copyright (c) 2014-2017 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
