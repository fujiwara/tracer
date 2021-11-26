# tracer

tracer is a tracing tool for Amazon ECS tasks.

tracer shows events and logs of the tasks order by timestamp.

### example

Run a task successfully and shutdown.

```
$ tracer default 84b5991528504856a2f003b7da9b2b82
2021-11-27T01:49:41.251+09:00   TASK    Created
2021-11-27T01:49:55.744+09:00   TASK    Pull started
2021-11-27T01:50:02.269+09:00   TASK    Pull stopped
2021-11-27T01:50:03.112+09:00   CONTAINER:nginx /docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
2021-11-27T01:50:03.112+09:00   CONTAINER:nginx /docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
2021-11-27T01:50:03.112+09:00   CONTAINER:nginx /docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
2021-11-27T01:50:03.123+09:00   CONTAINER:nginx 10-listen-on-ipv6-by-default.sh: info: Getting the checksum of /etc/nginx/conf.d/default.conf
2021-11-27T01:50:03.125+09:00   CONTAINER:nginx 10-listen-on-ipv6-by-default.sh: info: Enabled listen on IPv6 in /etc/nginx/conf.d/default.conf
2021-11-27T01:50:03.125+09:00   CONTAINER:nginx /docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
2021-11-27T01:50:03.128+09:00   CONTAINER:nginx /docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
2021-11-27T01:50:03.131+09:00   CONTAINER:nginx /docker-entrypoint.sh: Configuration complete; ready for start up
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: using the "epoll" event method
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: nginx/1.21.4
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: built by gcc 10.2.1 20210110 (Debian 10.2.1-6) 
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: OS: Linux 4.14.248-189.473.amzn2.aarch64
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1024:4096
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: start worker processes
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: start worker process 34
2021-11-27T01:50:03.135+09:00   CONTAINER:nginx 2021/11/26 16:50:03 [notice] 1#1: start worker process 35
2021-11-27T01:50:03.154+09:00   TASK    Started
2021-11-27T02:01:58.598+09:00   TASK    Stopping
2021-11-27T02:01:58.598+09:00   TASK    StoppedReason:Request stop task by user action.
2021-11-27T02:01:58.598+09:00   TASK    StoppedCode:UserInitiated
2021-11-27T02:01:59.129+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: signal 15 (SIGTERM) received, exiting
2021-11-27T02:01:59.129+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 34#34: exiting
2021-11-27T02:01:59.129+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 34#34: exit
2021-11-27T02:01:59.129+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 35#35: exiting
2021-11-27T02:01:59.129+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 35#35: exit
2021-11-27T02:01:59.179+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: signal 14 (SIGALRM) received
2021-11-27T02:01:59.215+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: signal 17 (SIGCHLD) received from 35
2021-11-27T02:01:59.215+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: worker process 35 exited with code 0
2021-11-27T02:01:59.215+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: signal 29 (SIGIO) received
2021-11-27T02:01:59.215+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: signal 17 (SIGCHLD) received from 34
2021-11-27T02:01:59.215+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: worker process 34 exited with code 0
2021-11-27T02:01:59.215+09:00   CONTAINER:nginx 2021/11/26 17:01:59 [notice] 1#1: exit
2021-11-27T02:02:10.402+09:00   TASK    Execution stopped
2021-11-27T02:02:33.559+09:00   TASK    Stopped
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
