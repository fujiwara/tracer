package tracer

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

const timeFormat = "2006-01-02T15:04:05.000Z07:00"

var epochBase = time.Unix(0, 0)

type Tracer struct {
	ctx      context.Context
	ecs      *ecs.ECS
	logs     *cloudwatchlogs.CloudWatchLogs
	timeline Timeline
}

func (t *Tracer) AddEvent(ts *time.Time, source, message string) {
	t.timeline.events = append(t.timeline.events, newEvent(ts, source, message))
}

type Timeline struct {
	events []*TimeLineEvent
	seen   map[string]bool
}

func (tl Timeline) Print(w io.Writer) {
	tls := make([]*TimeLineEvent, 0, len(tl.events))
	for _, e := range tl.events {
		if e.Timestamp == nil {
			continue
		}
		tls = append(tls, e)
	}
	sort.SliceStable(tls, func(i, j int) bool {
		return (*tls[i].Timestamp).Before(*tls[j].Timestamp)
	})
	for _, e := range tls {
		s := e.String()
		if !tl.seen[s] {
			fmt.Fprint(w, e.String())
			tl.seen[s] = true
		}
	}
}

type TimeLineEvent struct {
	Timestamp *time.Time
	Source    string
	Message   string
}

func (e *TimeLineEvent) String() string {
	ts := e.Timestamp.In(time.Local)
	return fmt.Sprintf("%s\t%s\t%s\n", ts.Format(timeFormat), e.Source, e.Message)
}

func New(ctx context.Context, sess *session.Session) (*Tracer, error) {
	tracer := &Tracer{
		ctx:      ctx,
		ecs:      ecs.New(sess),
		logs:     cloudwatchlogs.New(sess),
		timeline: Timeline{seen: make(map[string]bool)},
	}
	return tracer, nil
}

func newEvent(ts *time.Time, src, msg string) *TimeLineEvent {
	return &TimeLineEvent{
		Timestamp: ts,
		Source:    src,
		Message:   msg,
	}
}

func (t *Tracer) Run(cluster string, taskID string) error {
	task, err := t.traceTask(cluster, taskID)
	if err != nil {
		return err
	}
	if err := t.traceLogs(task); err != nil {
		return err
	}
	t.timeline.Print(os.Stdout)
	return nil
}

func (t *Tracer) traceTask(cluster string, taskID string) (*ecs.Task, error) {
	res, err := t.ecs.DescribeTasksWithContext(t.ctx, &ecs.DescribeTasksInput{
		Cluster: &cluster,
		Tasks:   []*string{&taskID},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to describe tasks")
	}
	if len(res.Tasks) == 0 {
		return nil, errors.New("no tasks found")
	}
	task := res.Tasks[0]
	t.AddEvent(task.CreatedAt, "TASK", "Created")
	t.AddEvent(task.StartedAt, "TASK", "Started")
	t.AddEvent(task.PullStartedAt, "TASK", "Pull started")
	t.AddEvent(task.PullStoppedAt, "TASK", "Pull stopped")
	t.AddEvent(task.StoppedAt, "TASK", "Stopped")
	t.AddEvent(task.StoppingAt, "TASK", "Stopping")
	if task.StoppedReason != nil {
		t.AddEvent(task.StoppingAt, "TASK", "StoppedReason:"+*task.StoppedReason)
	}
	if task.StopCode != nil {
		t.AddEvent(task.StoppingAt, "TASK", "StoppedCode:"+*task.StopCode)
	}
	t.AddEvent(task.ExecutionStoppedAt, "TASK", "Execution stopped")
	return task, nil
}

func (t *Tracer) traceLogs(task *ecs.Task) error {
	res, err := t.ecs.DescribeTaskDefinitionWithContext(t.ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task.TaskDefinitionArn,
	})
	if err != nil {
		return errors.Wrap(err, "failed to describe task definition")
	}
	for _, c := range res.TaskDefinition.ContainerDefinitions {
		if c.LogConfiguration == nil {
			continue
		}
		if aws.StringValue(c.LogConfiguration.LogDriver) != "awslogs" {
			continue
		}
		opt := c.LogConfiguration.Options
		logGroup := *opt["awslogs-group"]
		logStream := *opt["awslogs-stream-prefix"] + "/" + *c.Name + "/" + taskID(task)
		t.followLogs(*c.Name, logGroup, logStream, task.CreatedAt, task.StoppingAt)
	}
	return nil
}

func taskID(task *ecs.Task) string {
	an := aws.StringValue(task.TaskArn)
	return an[strings.LastIndex(an, "/")+1:]
}

func (t *Tracer) followLogs(containerName, group, stream string, start, end *time.Time) error {
	res, err := t.logs.GetLogEventsWithContext(t.ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		StartTime:     aws.Int64(timeToInt64msec(*start)),
		Limit:         aws.Int64(1000),
	})
	if err != nil {
		return err
	}
	for _, e := range res.Events {
		ts := msecToTime(aws.Int64Value(e.Timestamp))
		t.AddEvent(&ts, "CONTAINER:"+containerName, aws.StringValue(e.Message))
	}
	if end == nil {
		return nil
	}

	res, err = t.logs.GetLogEventsWithContext(t.ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		StartTime:     aws.Int64(timeToInt64msec(*end) - 300*1000),
		Limit:         aws.Int64(1000),
	})
	if err != nil {
		return err
	}
	for _, e := range res.Events {
		ts := msecToTime(aws.Int64Value(e.Timestamp))
		t.AddEvent(&ts, "CONTAINER:"+containerName, aws.StringValue(e.Message))
	}
	return nil
}

func msecToTime(i int64) time.Time {
	return epochBase.Add(time.Duration(i) * time.Millisecond)
}

func timeToInt64msec(t time.Time) int64 {
	return int64(t.Sub(epochBase) / time.Millisecond)
}
