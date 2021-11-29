# tracer

tracer is a tracing tool for Amazon ECS tasks.

tracer shows events and logs of the tasks order by timestamp.

## Install

```console
$ brew install fujiwara/tap/tracer
```

[Binary Releases](https://github.com/fujiwara/tracer/releases)

## Example

Run a task successfully and shutdown.

```
2021-11-29T10:42:33.521+09:00	TASK	Created
2021-11-29T10:42:37.103+09:00	TASK	Connected
2021-11-29T10:42:48.298+09:00	TASK	Pull started
2021-11-29T10:42:55.473+09:00	TASK	Pull stopped
2021-11-29T10:42:56.374+09:00	CONTAINER:nginx	/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
2021-11-29T10:42:56.374+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
2021-11-29T10:42:56.375+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
2021-11-29T10:42:56.380+09:00	CONTAINER:nginx	10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
2021-11-29T10:42:56.418+09:00	TASK	Started
2021-11-29T10:42:56.446+09:00	CONTAINER:nginx	10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
2021-11-29T10:42:56.446+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
2021-11-29T10:42:56.448+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
2021-11-29T10:42:56.449+09:00	CONTAINER:nginx	/docker-entrypoint.sh: Configuration complete; ready for start up
2021-11-29T10:42:56.452+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: using the "epoll" event method
2021-11-29T10:42:56.452+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: nginx/1.21.4
2021-11-29T10:42:56.452+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6)
2021-11-29T10:42:56.452+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: OS: Linux 4.14.248-189.473.amzn2.aarch64
2021-11-29T10:42:56.452+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1024:4096
2021-11-29T10:42:56.452+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: start worker processes
2021-11-29T10:42:56.453+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: start worker process 31
2021-11-29T10:42:56.453+09:00	CONTAINER:nginx	2021/11/29 01:42:56 [notice] 1#1: start worker process 32
2021-11-29T10:43:12.913+09:00	CONTAINER:nginx	10.3.3.10 - - [29/Nov/2021:01:43:12 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:43:42.898+09:00	CONTAINER:nginx	10.3.3.10 - - [29/Nov/2021:01:43:42 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:44:09.172+09:00	CONTAINER:nginx	10.3.1.178 - - [29/Nov/2021:01:44:09 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:44:12.929+09:00	CONTAINER:nginx	10.3.3.10 - - [29/Nov/2021:01:44:12 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:44:39.197+09:00	CONTAINER:nginx	10.3.1.178 - - [29/Nov/2021:01:44:39 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:44:42.960+09:00	CONTAINER:nginx	10.3.3.10 - - [29/Nov/2021:01:44:42 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:45:09.227+09:00	CONTAINER:nginx	10.3.1.178 - - [29/Nov/2021:01:45:09 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:45:12.966+09:00	CONTAINER:nginx	10.3.3.10 - - [29/Nov/2021:01:45:12 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:45:39.237+09:00	CONTAINER:nginx	10.3.1.178 - - [29/Nov/2021:01:45:39 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:45:42.973+09:00	CONTAINER:nginx	10.3.3.10 - - [29/Nov/2021:01:45:42 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:46:09.225+09:00	CONTAINER:nginx	10.3.1.178 - - [29/Nov/2021:01:46:09 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:46:13.004+09:00	CONTAINER:nginx	10.3.3.10 - - [29/Nov/2021:01:46:13 +0000] "GET / HTTP/1.1" 200 615 "-" "ELB-HealthChecker/2.0" "-"
2021-11-29T10:46:30.906+09:00	TASK	Stopping
2021-11-29T10:46:30.906+09:00	TASK	StoppedReason:Scaling activity initiated by (deployment ecs-svc/8709920613704280865)
2021-11-29T10:46:30.906+09:00	TASK	StoppedCode:ServiceSchedulerInitiated
2021-11-29T10:46:31.937+09:00	CONTAINER:nginx	2021/11/29 01:46:31 [notice] 1#1: signal 15 (SIGTERM) received, exiting
2021-11-29T10:46:31.938+09:00	CONTAINER:nginx	2021/11/29 01:46:31 [notice] 32#32: exiting
2021-11-29T10:46:31.938+09:00	CONTAINER:nginx	2021/11/29 01:46:31 [notice] 32#32: exit
2021-11-29T10:46:31.938+09:00	CONTAINER:nginx	2021/11/29 01:46:31 [notice] 31#31: exiting
2021-11-29T10:46:31.938+09:00	CONTAINER:nginx	2021/11/29 01:46:31 [notice] 31#31: exit
2021-11-29T10:46:31.988+09:00	CONTAINER:nginx	2021/11/29 01:46:31 [notice] 1#1: signal 14 (SIGALRM) received
2021-11-29T10:46:32.035+09:00	CONTAINER:nginx	2021/11/29 01:46:32 [notice] 1#1: signal 17 (SIGCHLD) received from 32
2021-11-29T10:46:32.035+09:00	CONTAINER:nginx	2021/11/29 01:46:32 [notice] 1#1: worker process 32 exited with code 0
2021-11-29T10:46:32.035+09:00	CONTAINER:nginx	2021/11/29 01:46:32 [notice] 1#1: signal 29 (SIGIO) received
2021-11-29T10:46:32.035+09:00	CONTAINER:nginx	2021/11/29 01:46:32 [notice] 1#1: signal 17 (SIGCHLD) received from 31
2021-11-29T10:46:32.035+09:00	CONTAINER:nginx	2021/11/29 01:46:32 [notice] 1#1: worker process 31 exited with code 0
2021-11-29T10:46:32.035+09:00	CONTAINER:nginx	2021/11/29 01:46:32 [notice] 1#1: exit
2021-11-29T10:46:43.234+09:00	TASK	Execution stopped
2021-11-29T10:47:06.292+09:00	TASK	Stopped
2021-11-29T10:47:06.292+09:00	CONTAINER:nginx	STOPPED (exit code: 0)
```

Failed to run task. (typo container image URL)

```
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
