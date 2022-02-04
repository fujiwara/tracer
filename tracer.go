package tracer

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
)

const timeFormat = "2006-01-02T15:04:05.000Z07:00"

var epochBase = time.Unix(0, 0)

var MaxFetchLogs = 100

type Tracer struct {
	ctx      context.Context
	ecs      *ecs.ECS
	logs     *cloudwatchlogs.CloudWatchLogs
	timeline *Timeline
	Duration time.Duration

	now       time.Time
	headBegin time.Time
	headEnd   time.Time
	tailBegin time.Time
	tailEnd   time.Time
}

func (t *Tracer) AddEvent(ts *time.Time, source, message string) {
	t.timeline.Add(newEvent(ts, source, message))
}

type Timeline struct {
	events []*TimeLineEvent
	seen   map[string]bool
	mu     sync.Mutex
}

func (tl *Timeline) Add(event *TimeLineEvent) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.events = append(tl.events, event)
}

func (tl *Timeline) Print(w io.Writer) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

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

func NewWithSession(ctx context.Context, sess *session.Session) (*Tracer, error) {
	return New(ctx, ecs.New(sess), cloudwatchlogs.New(sess))
}

func New(ctx context.Context, ecsSv *ecs.ECS, logsSv *cloudwatchlogs.CloudWatchLogs) (*Tracer, error) {
	return &Tracer{
		ctx:  ctx,
		ecs:  ecsSv,
		logs: logsSv,
		timeline: &Timeline{
			seen: make(map[string]bool),
		},
		Duration: time.Minute,
	}, nil
}

func newEvent(ts *time.Time, src, msg string) *TimeLineEvent {
	return &TimeLineEvent{
		Timestamp: ts,
		Source:    src,
		Message:   msg,
	}
}

func (t *Tracer) Run(cluster string, taskID string) error {
	t.now = time.Now()

	if cluster == "" {
		return t.listClusters()
	}

	if taskID == "" {
		return t.listAllTasks(cluster)
	}

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

	t.setBoundaries(task)

	taskGroup := strings.SplitN(aws.StringValue(task.Group), ":", 2)
	if len(taskGroup) == 2 && taskGroup[0] == "service" {
		t.fetchServiceEvents(cluster, taskGroup[1])
	}

	t.AddEvent(task.CreatedAt, "TASK", "Created")
	t.AddEvent(task.ConnectivityAt, "TASK", "Connected")
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

	for _, c := range task.Containers {
		containerName := *c.Name
		msg := fmt.Sprintf(*c.LastStatus)
		if c.ExitCode != nil {
			msg += fmt.Sprintf(" (exit code: %d)", *c.ExitCode)
		}
		if c.Reason != nil {
			msg += fmt.Sprintf(" (reason: %s)", *c.Reason)
		}
		var ts *time.Time
		if aws.StringValue(c.LastStatus) == "RUNNING" {
			ts = &t.now
		} else {
			ts = task.StoppedAt
		}
		t.AddEvent(ts, "CONTAINER:"+containerName, msg)
	}

	return task, nil
}

func (t *Tracer) traceLogs(task *ecs.Task) error {
	res, err := t.ecs.DescribeTaskDefinitionWithContext(t.ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task.TaskDefinitionArn,
	})
	if err != nil {
		return errors.Wrap(err, "failed to describe task definition")
	}
	var wg sync.WaitGroup
	for _, c := range res.TaskDefinition.ContainerDefinitions {
		containerName := *c.Name
		if c.LogConfiguration == nil {
			continue
		}
		if aws.StringValue(c.LogConfiguration.LogDriver) != "awslogs" {
			continue
		}
		opt := c.LogConfiguration.Options
		logGroup := *opt["awslogs-group"]
		logStream := *opt["awslogs-stream-prefix"] + "/" + *c.Name + "/" + taskID(task)
		wg.Add(1)
		go func() {
			defer wg.Done()
			// head of logs
			t.fetchLogs(containerName, logGroup, logStream, &t.headBegin, &t.headEnd)

			// tail of logs
			t.fetchLogs(containerName, logGroup, logStream, &t.tailBegin, nil)
		}()
	}
	wg.Wait()
	return nil
}

func taskID(task *ecs.Task) string {
	an := aws.StringValue(task.TaskArn)
	return an[strings.LastIndex(an, "/")+1:]
}

func (t *Tracer) fetchServiceEvents(cluster, service string) error {
	res, err := t.ecs.DescribeServicesWithContext(t.ctx, &ecs.DescribeServicesInput{
		Cluster:  &cluster,
		Services: []*string{&service},
	})
	if err != nil {
		return errors.Wrap(err, "failed to describe services")
	}
	if len(res.Services) == 0 {
		return errors.New("no services found")
	}
	for _, e := range res.Services[0].Events {
		ts := *e.CreatedAt
		if ts.After(t.headBegin) && ts.Before(t.headEnd) || ts.After(t.tailBegin) && ts.Before(t.tailEnd) {
			t.AddEvent(e.CreatedAt, "SERVICE", *e.Message)
		}
	}
	return nil
}

func (t *Tracer) fetchLogs(containerName, group, stream string, from, to *time.Time) error {
	var nextToken *string
	in := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		Limit:         aws.Int64(100),
	}
	if from != nil {
		in.StartTime = aws.Int64(timeToInt64msec(*from))
	} else {
		in.StartFromHead = aws.Bool(true)
	}
	if to != nil {
		in.EndTime = aws.Int64(timeToInt64msec(*to))
	}

	fetched := 0
	for {
		if nextToken != nil {
			in.NextToken = nextToken
			in.StartFromHead = nil
		}
		res, err := t.logs.GetLogEventsWithContext(t.ctx, in)
		if err != nil {
			return err
		}
		fetched++
		for _, e := range res.Events {
			ts := msecToTime(aws.Int64Value(e.Timestamp))
			t.AddEvent(&ts, "CONTAINER:"+containerName, aws.StringValue(e.Message))
		}
		if aws.StringValue(nextToken) == aws.StringValue(res.NextForwardToken) {
			break
		}
		if fetched >= MaxFetchLogs {
			break
		}
		nextToken = res.NextForwardToken
	}
	return nil
}

func (t *Tracer) listAllTasks(cluster string) error {
	for _, s := range []string{"RUNNING", "PENDING", "STOPPED"} {
		err := t.listTasks(cluster, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Tracer) listClusters() error {
	res, err := t.ecs.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		return err
	}
	clusters := make([]string, 0, len(res.ClusterArns))
	for _, c := range res.ClusterArns {
		clusters = append(clusters, arnToName(aws.StringValue(c)))
	}
	sort.Strings(clusters)
	fmt.Println(strings.Join(clusters, "\n"))
	return nil
}

func (t *Tracer) listTasks(cluster, status string) error {
	var nextToken *string
	for {
		listRes, err := t.ecs.ListTasksWithContext(t.ctx, &ecs.ListTasksInput{
			Cluster:       &cluster,
			DesiredStatus: aws.String(status),
			NextToken:     nextToken,
		})
		if err != nil {
			return errors.Wrap(err, "failed to list tasks")
		}
		if len(listRes.TaskArns) == 0 {
			break
		}
		res, err := t.ecs.DescribeTasksWithContext(t.ctx, &ecs.DescribeTasksInput{
			Cluster: &cluster,
			Tasks:   listRes.TaskArns,
		})
		if err != nil {
			return errors.Wrap(err, "failed to describe tasks")
		}
		for _, task := range res.Tasks {
			fmt.Println(strings.Join(taskToColumns(task), "\t"))
		}
		if nextToken = listRes.NextToken; nextToken == nil {
			break
		}
	}
	return nil
}

func (t *Tracer) setBoundaries(task *ecs.Task) {
	t.headBegin = task.CreatedAt.Add(-t.Duration)
	if task.StartedAt != nil {
		t.headEnd = task.StartedAt.Add(t.Duration)
	} else {
		t.headEnd = task.CreatedAt.Add(t.Duration)
	}

	if task.StoppingAt != nil {
		t.tailBegin = task.StoppingAt.Add(-t.Duration)
	} else {
		t.tailBegin = t.now.Add(-t.Duration)
	}
	if task.StoppedAt != nil {
		t.tailEnd = task.StoppedAt.Add(t.Duration)
	} else {
		t.tailEnd = t.now
	}
}

func msecToTime(i int64) time.Time {
	return epochBase.Add(time.Duration(i) * time.Millisecond)
}

func timeToInt64msec(t time.Time) int64 {
	return int64(t.Sub(epochBase) / time.Millisecond)
}

func arnToName(arn string) string {
	return arn[strings.LastIndex(arn, "/")+1:]
}

func taskToColumns(task *ecs.Task) []string {
	return []string{
		arnToName(*task.TaskArn),
		arnToName(*task.TaskDefinitionArn),
		aws.StringValue(task.LastStatus),
		aws.StringValue(task.DesiredStatus),
		task.CreatedAt.In(time.Local).Format(time.RFC3339),
		aws.StringValue(task.Group),
		aws.StringValue(task.LaunchType),
	}
}
