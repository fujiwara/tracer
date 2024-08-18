# tracer

tracer is a tracing tool for Amazon ECS tasks.

tracer shows events and logs of the tasks order by timestamp.

## Install

```console
$ brew install fujiwara/tap/tracer
```

[Binary Releases](https://github.com/fujiwara/tracer/releases)

## Usage

### as a CLI

```
Usage of tracer:
tracer [options] [cluster] [task-id]

  -duration duration
        fetch logs duration from created / before stopping (default 1m0s)
  -json
        output as JSON lines
  -sns string
        SNS topic ARN
  -stdout
        output to stdout (default true)
  -version
        show the version
```

Environment variable `AWS_REGION` is required.

- `tracer` (no arguments) shows list of clusters.
- `tracer {cluster}` shows all tasks in the clusters.
- `tracer {cluster} {task-id}` shows a tracing logs of the task.


### as a Lambda function

`tracer` also runs on AWS Lambda functions invoked by EventBridge's "ECS Task State Change" events.

1. Put a `tracer` binary into a lambda function's archive(zip) as `bootstrap` named.
1. Set to call the lambda function by EvnetBridge rule as below.
   ```json
   {
     "source": ["aws.ecs"],
     "detail-type": ["ECS Task State Change"]
   }
   ```
1. The tracer lambda function will put trace logs when ECS tasks STOPPED.

See also [lambda directory](lambda/).

### IAM permissions

tracer requires IAM permissions as below.

- `ecs:Describe*`
- `ecs:List*`
- `logs:GetLog*`

See also [example.tf](lambda/example.tf).


## Example

Run a task successfully and shutdown.

```console
$ tracer default 834a5628bef14f2dbb81c7bc0b272160
2021-12-03T11:06:21.633+09:00	TASK	Created
2021-12-03T11:06:21.664+09:00	SERVICE	(service nginx-local) has started 1 tasks: (task 834a5628bef14f2dbb81c7bc0b272160).
2021-12-03T11:06:22.342+09:00	SERVICE	(service nginx-local) was unable to place a task. Reason: Capacity is unavailable at this time. Please try again later or in a different availability zone. For more information, see the Troubleshooting section of the Amazon ECS Developer Guide.
2021-12-03T11:06:24.906+09:00	TASK	Connected
2021-12-03T11:06:39.602+09:00	TASK	Pull started
2021-12-03T11:06:46.366+09:00	TASK	Pull stopped
2021-12-03T11:06:46.746+09:00	CONTAINER:nginx	/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
2021-12-03T11:06:46.746+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
2021-12-03T11:06:46.746+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
2021-12-03T11:06:46.758+09:00	CONTAINER:nginx	10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
2021-12-03T11:06:46.762+09:00	CONTAINER:nginx	10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
2021-12-03T11:06:46.762+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
2021-12-03T11:06:46.768+09:00	TASK	Started
2021-12-03T11:06:46.820+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
2021-12-03T11:06:46.832+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: using the "epoll" event method
2021-12-03T11:06:46.832+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: nginx/1.21.4
2021-12-03T11:06:46.832+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6) 
2021-12-03T11:06:46.832+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: OS: Linux 4.14.248-189.473.amzn2.aarch64
2021-12-03T11:06:46.832+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1024:4096
2021-12-03T11:06:46.832+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Configuration complete; ready for start up
2021-12-03T11:06:46.832+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: start worker processes
2021-12-03T11:06:46.837+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: start worker process 37
2021-12-03T11:06:46.837+09:00	CONTAINER:nginx	2021/12/03 02:06:46 [notice] 1#1: start worker process 38
2021-12-03T11:21:36.818+09:00	CONTAINER:nginx	10.3.1.18 - - [03/Dec/2021:02:21:36 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:21:36.836+09:00	CONTAINER:nginx	10.3.3.10 - - [03/Dec/2021:02:21:36 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:21:46.819+09:00	CONTAINER:nginx	10.3.1.18 - - [03/Dec/2021:02:21:46 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:21:46.837+09:00	CONTAINER:nginx	10.3.3.10 - - [03/Dec/2021:02:21:46 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:21:56.820+09:00	CONTAINER:nginx	10.3.1.18 - - [03/Dec/2021:02:21:56 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:21:56.839+09:00	CONTAINER:nginx	10.3.3.10 - - [03/Dec/2021:02:21:56 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:22:06.821+09:00	CONTAINER:nginx	10.3.1.18 - - [03/Dec/2021:02:22:06 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:22:06.840+09:00	CONTAINER:nginx	10.3.3.10 - - [03/Dec/2021:02:22:06 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:22:12.681+09:00	CONTAINER:nginx	10.3.1.18 - - [03/Dec/2021:02:22:12 +0000] "GET / HTTP/1.1" 200 615 "-" "Mozilla/5.0 (compatible; Nimbostratus-Bot/v1.3.2; http://cloudsystemnetworks.com)" "209.17.96.194"
2021-12-03T11:22:16.821+09:00	CONTAINER:nginx	10.3.1.18 - - [03/Dec/2021:02:22:16 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:22:16.841+09:00	CONTAINER:nginx	10.3.3.10 - - [03/Dec/2021:02:22:16 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:22:19.833+09:00	SERVICE	(service nginx-local) deregistered 1 targets in (target-group arn:aws:elasticloadbalancing:ap-northeast-1:314472643515:targetgroup/alpha/6a301850702273d9)
2021-12-03T11:22:19.834+09:00	SERVICE	(service nginx-local) has begun draining connections on 1 tasks.
2021-12-03T11:22:26.822+09:00	CONTAINER:nginx	10.3.1.18 - - [03/Dec/2021:02:22:26 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:22:26.842+09:00	CONTAINER:nginx	10.3.3.10 - - [03/Dec/2021:02:22:26 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-12-03T11:22:28.910+09:00	TASK	Stopping
2021-12-03T11:22:28.910+09:00	TASK	StoppedReason:Scaling activity initiated by (deployment ecs-svc/8709920613704280865)
2021-12-03T11:22:28.910+09:00	TASK	StoppedCode:ServiceSchedulerInitiated
2021-12-03T11:22:28.938+09:00	SERVICE	(service nginx-local) has stopped 1 running tasks: (task 834a5628bef14f2dbb81c7bc0b272160).
2021-12-03T11:22:29.244+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: signal 15 (SIGTERM) received, exiting
2021-12-03T11:22:29.245+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 37#37: exiting
2021-12-03T11:22:29.245+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 37#37: exit
2021-12-03T11:22:29.245+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 38#38: exiting
2021-12-03T11:22:29.245+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 38#38: exit
2021-12-03T11:22:29.294+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: signal 14 (SIGALRM) received
2021-12-03T11:22:29.328+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: signal 17 (SIGCHLD) received from 37
2021-12-03T11:22:29.328+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: worker process 37 exited with code 0
2021-12-03T11:22:29.328+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: signal 29 (SIGIO) received
2021-12-03T11:22:29.329+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: signal 17 (SIGCHLD) received from 38
2021-12-03T11:22:29.329+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: worker process 38 exited with code 0
2021-12-03T11:22:29.329+09:00	CONTAINER:nginx	2021/12/03 02:22:29 [notice] 1#1: exit
2021-12-03T11:22:38.224+09:00	SERVICE	(service nginx-local) has reached a steady state.
2021-12-03T11:22:40.527+09:00	TASK	Execution stopped
2021-12-03T11:23:04.873+09:00	TASK	Stopped
2021-12-03T11:23:04.873+09:00	CONTAINER:nginx	STOPPED (exit code: 0)
```

Failed to run task. (typo container image URL)

```console
$ tracer default 9f654c76cde14c7c85cf54dce087658a
2021-11-27T02:29:15.055+09:00   TASK    Created
2021-11-27T02:29:33.527+09:00   TASK    Execution stopped
2021-11-27T02:29:43.569+09:00   TASK    Stopping
2021-11-27T02:29:43.569+09:00   TASK    StoppedReason:CannotPullContainerError: inspect image has been retried 1 time(s): failed to resolve ref "docker.io/library/ngin:latest": pull access denied, repository does not exist or may require authorization: server message: insufficient_scope: authorization failed
2021-11-27T02:29:43.569+09:00   TASK    StoppedCode:TaskFailedToStart
2021-11-27T02:29:57.070+09:00   TASK    Stopped
```

## LICENSE

MIT

## Author

fujiwara
