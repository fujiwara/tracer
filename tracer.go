package tracer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

const (
	snsMaxPayloadSize = 256 * 1024
)

var TimeFormat = "2006-01-02T15:04:05.000Z07:00"

var epochBase = time.Unix(0, 0)

var MaxFetchLogs = 100

type Tracer struct {
	ecs      *ecs.Client
	logs     *cloudwatchlogs.Client
	sns      *sns.Client
	timeline *Timeline
	buf      *bytes.Buffer
	w        io.Writer

	now       time.Time
	headBegin time.Time
	headEnd   time.Time
	tailBegin time.Time
	tailEnd   time.Time

	option *RunOption
}

func (t *Tracer) AddEvent(ts time.Time, source, message string) {
	t.timeline.Add(newEvent(ts, source, message))
}

func NewTimeline() *Timeline {
	return &Timeline{
		seen: make(map[string]bool),
	}
}

type Timeline struct {
	events []*TimelineEvent
	seen   map[string]bool
	mu     sync.Mutex
}

func (tl *Timeline) Add(event *TimelineEvent) {
	if event.Timestamp.IsZero() { // ignore zero time event
		return
	}
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.events = append(tl.events, event)
}

func (tl *Timeline) Print(w io.Writer, json bool) (int, error) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tls := make([]*TimelineEvent, 0, len(tl.events))
	tls = append(tls, tl.events...)
	sort.SliceStable(tls, func(i, j int) bool {
		return tls[i].Timestamp.Before(tls[j].Timestamp)
	})
	n := 0
	toString := func(e *TimelineEvent) string {
		if json {
			return e.JSON()
		}
		return e.String()
	}
	for _, e := range tls {
		s := toString(e)
		if tl.seen[s] {
			continue
		}
		l, err := fmt.Fprint(w, s)
		if err != nil {
			return n, err
		}
		n += l
		tl.seen[s] = true
	}
	return n, nil
}

type TimelineEvent struct {
	Timestamp time.Time
	Source    string
	Message   string
}

func (e *TimelineEvent) String() string {
	ts := e.Timestamp.In(time.Local)
	return fmt.Sprintf("%s\t%s\t%s\n", ts.Format(TimeFormat), e.Source, e.Message)
}

func (e *TimelineEvent) JSON() string {
	ts := e.Timestamp.In(time.Local)
	b, _ := json.Marshal(struct {
		Time    string `json:"time"`
		Source  string `json:"src"`
		Message string `json:"msg"`
	}{
		Time:    ts.Format(TimeFormat),
		Source:  e.Source,
		Message: e.Message,
	})
	return string(b) + "\n"
}

func New(ctx context.Context) (*Tracer, error) {
	region := os.Getenv("AWS_REGION")
	awscfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return NewWithConfig(awscfg)
}

func NewWithConfig(config aws.Config) (*Tracer, error) {
	return &Tracer{
		ecs:      ecs.NewFromConfig(config),
		logs:     cloudwatchlogs.NewFromConfig(config),
		sns:      sns.NewFromConfig(config),
		timeline: NewTimeline(),
		buf:      new(bytes.Buffer),
		w:        os.Stdout,
	}, nil
}

func newEvent(ts time.Time, src, msg string) *TimelineEvent {
	return &TimelineEvent{
		Timestamp: ts,
		Source:    src,
		Message:   msg,
	}
}

type RunOption struct {
	Stdout      bool
	SNSTopicArn string
	Duration    time.Duration
	JSON        bool
}

func (t *Tracer) SetOutput(w io.Writer) {
	t.w = w
}

func (t *Tracer) Run(ctx context.Context, cluster string, taskID string, opt *RunOption) error {
	t.now = time.Now()
	t.option = opt

	defer func() { t.report(ctx, cluster, taskID) }()

	if cluster == "" {
		return t.listClusters(ctx, opt)
	}

	if taskID == "" {
		return t.listAllTasks(ctx, cluster)
	}

	task, err := t.traceTask(ctx, cluster, taskID)
	if err != nil {
		return err
	}

	defer func() {
		if _, err := t.timeline.Print(t.buf, opt.JSON); err != nil {
			slog.Error("failed to print timeline", "error", err)
		}
	}()
	if err := t.traceLogs(ctx, task); err != nil {
		return err
	}

	return nil
}

func (t *Tracer) report(ctx context.Context, cluster, taskID string) {
	opt := t.option
	if opt.Stdout {
		sub := &subject{cluster, taskID}
		if opt.JSON {
			fmt.Fprintln(t.w, sub.JSON())
		} else {
			fmt.Fprintln(t.w, sub.String())
		}
		if _, err := t.WriteTo(t.w); err != nil {
			slog.Error("failed to write to output", "error", err)
		}
	}
	if opt.SNSTopicArn != "" {
		if err := t.Publish(ctx, opt.SNSTopicArn, cluster, taskID); err != nil {
			slog.Error("failed to publish to SNS", "error", err)
		}
	}
}

func (t *Tracer) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, t.buf.String())
	return int64(n), err
}

type subject struct {
	Cluster string `json:"cluster"`
	TaskID  string `json:"task_id"`
}

func (s *subject) JSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s *subject) String() string {
	str := "Tracer:"
	if s.TaskID != "" {
		str += " " + s.TaskID
	} else if s.Cluster != "" {
		str += " tasks"
	}
	if s.Cluster != "" {
		str += " on " + s.Cluster
	} else {
		str += " clusters"
	}
	return str
}

const (
	snsSubjectLimitLength = 100
	ellipsisString        = "..."
)

func (t *Tracer) Publish(ctx context.Context, topicArn, cluster, taskID string) error {
	msg := t.buf.String()
	if len(msg) >= snsMaxPayloadSize {
		msg = msg[:snsMaxPayloadSize]
	}

	s := (&subject{cluster, taskID}).String()
	if len(s) > snsSubjectLimitLength {
		s = s[0:snsSubjectLimitLength-len(ellipsisString)] + ellipsisString
	}
	_, err := t.sns.Publish(ctx, &sns.PublishInput{
		Message:  &msg,
		Subject:  aws.String(s),
		TopicArn: &topicArn,
	})
	return err
}

func (t *Tracer) traceTask(ctx context.Context, cluster string, taskID string) (*ecsTypes.Task, error) {
	res, err := t.ecs.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: &cluster,
		Tasks:   []string{taskID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe tasks: %w", err)
	}
	if len(res.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks found. cluster: %s, task_id: %s", cluster, taskID)
	}
	task := res.Tasks[0]

	t.setBoundaries(&task)

	taskGroup := strings.SplitN(aws.ToString(task.Group), ":", 2)
	if len(taskGroup) == 2 && taskGroup[0] == "service" {
		t.fetchServiceEvents(ctx, cluster, taskGroup[1])
	}

	t.AddEvent(aws.ToTime(task.CreatedAt), "TASK", "Created")
	t.AddEvent(aws.ToTime(task.ConnectivityAt), "TASK", "Connected")
	t.AddEvent(aws.ToTime(task.StartedAt), "TASK", "Started")
	t.AddEvent(aws.ToTime(task.PullStartedAt), "TASK", "Pull started")
	t.AddEvent(aws.ToTime(task.PullStoppedAt), "TASK", "Pull stopped")
	t.AddEvent(aws.ToTime(task.StoppedAt), "TASK", "Stopped")
	t.AddEvent(aws.ToTime(task.StoppingAt), "TASK", "Stopping")
	if task.StoppedReason != nil {
		t.AddEvent(aws.ToTime(task.StoppingAt), "TASK", "StoppedReason:"+aws.ToString(task.StoppedReason))
	}
	t.AddEvent(aws.ToTime(task.StoppingAt), "TASK", "StoppedCode:"+string(task.StopCode))
	t.AddEvent(aws.ToTime(task.ExecutionStoppedAt), "TASK", "Execution stopped")

	for _, c := range task.Containers {
		containerName := *c.Name
		msg := fmt.Sprintf("LastStatus:%s HealthStatus:%s", *c.LastStatus, c.HealthStatus)
		if c.ExitCode != nil {
			msg += fmt.Sprintf(" (exit code: %d)", *c.ExitCode)
		}
		if c.Reason != nil {
			msg += fmt.Sprintf(" (reason: %s)", *c.Reason)
		}
		t.AddEvent(t.now, "CONTAINER:"+containerName, msg)
	}

	t.AddEvent(t.now, "TASK", "LastStatus:"+aws.ToString(task.LastStatus))

	return &task, nil
}

func (t *Tracer) traceLogs(ctx context.Context, task *ecsTypes.Task) error {
	res, err := t.ecs.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task.TaskDefinitionArn,
	})
	if err != nil {
		return fmt.Errorf("failed to describe task definition: %w", err)
	}
	var wg sync.WaitGroup
	for _, c := range res.TaskDefinition.ContainerDefinitions {
		containerName := *c.Name
		if c.LogConfiguration == nil {
			continue
		}
		if c.LogConfiguration.LogDriver != ecsTypes.LogDriverAwslogs {
			continue
		}
		opt := c.LogConfiguration.Options
		logGroup := opt["awslogs-group"]
		logStream := opt["awslogs-stream-prefix"] + "/" + *c.Name + "/" + taskID(task)
		wg.Add(1)
		go func() {
			defer wg.Done()
			// head of logs
			t.fetchLogs(ctx, containerName, logGroup, logStream, &t.headBegin, &t.headEnd)

			// tail of logs
			t.fetchLogs(ctx, containerName, logGroup, logStream, &t.tailBegin, nil)
		}()
	}
	wg.Wait()
	return nil
}

func taskID(task *ecsTypes.Task) string {
	an := aws.ToString(task.TaskArn)
	return an[strings.LastIndex(an, "/")+1:]
}

func (t *Tracer) fetchServiceEvents(ctx context.Context, cluster, service string) error {
	res, err := t.ecs.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  &cluster,
		Services: []string{service},
	})
	if err != nil {
		return fmt.Errorf("failed to describe services: %w", err)
	}
	if len(res.Services) == 0 {
		return fmt.Errorf("no services found: %w", err)
	}
	for _, e := range res.Services[0].Events {
		ts := aws.ToTime(e.CreatedAt)
		if ts.After(t.headBegin) && ts.Before(t.headEnd) || ts.After(t.tailBegin) && ts.Before(t.tailEnd) {
			t.AddEvent(ts, "SERVICE", aws.ToString(e.Message))
		}
	}
	return nil
}

func (t *Tracer) fetchLogs(ctx context.Context, containerName, group, stream string, from, to *time.Time) error {
	var nextToken *string
	in := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		Limit:         aws.Int32(100),
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
		res, err := t.logs.GetLogEvents(ctx, in)
		if err != nil {
			return err
		}
		fetched++
		for _, e := range res.Events {
			ts := msecToTime(aws.ToInt64(e.Timestamp))
			t.AddEvent(ts, "CONTAINER:"+containerName, aws.ToString(e.Message))
		}
		if aws.ToString(nextToken) == aws.ToString(res.NextForwardToken) {
			break
		}
		if fetched >= MaxFetchLogs {
			break
		}
		nextToken = res.NextForwardToken
	}
	return nil
}

func (t *Tracer) listAllTasks(ctx context.Context, cluster string) error {
	for _, s := range []ecsTypes.DesiredStatus{
		ecsTypes.DesiredStatusRunning,
		ecsTypes.DesiredStatusPending,
		ecsTypes.DesiredStatusStopped,
	} {
		err := t.listTasks(ctx, cluster, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Tracer) listClusters(ctx context.Context, opt *RunOption) error {
	res, err := t.ecs.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return err
	}
	clusters := make([]string, 0, len(res.ClusterArns))
	for _, c := range res.ClusterArns {
		clusters = append(clusters, arnToName(c))
	}
	sort.Strings(clusters)
	if opt.JSON {
		err := json.NewEncoder(t.buf).Encode(
			struct {
				Clusters []string `json:"clusters"`
			}{clusters},
		)
		if err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
		return nil
	}
	for _, c := range clusters {
		t.buf.WriteString(c)
		t.buf.WriteByte('\n')
	}
	return nil
}

func (t *Tracer) listTasks(ctx context.Context, cluster string, status ecsTypes.DesiredStatus) error {
	var nextToken *string
	for {
		listRes, err := t.ecs.ListTasks(ctx, &ecs.ListTasksInput{
			Cluster:       &cluster,
			DesiredStatus: status,
			NextToken:     nextToken,
		})
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}
		if len(listRes.TaskArns) == 0 {
			break
		}
		res, err := t.ecs.DescribeTasks(ctx, &ecs.DescribeTasksInput{
			Cluster: &cluster,
			Tasks:   listRes.TaskArns,
		})
		if err != nil {
			return fmt.Errorf("failed to describe tasks: %w", err)
		}
		for _, ts := range res.Tasks {
			task := newTask(&ts)
			if t.option.JSON {
				t.buf.WriteString(task.JSON())
			} else {
				t.buf.WriteString(task.String())
			}
		}
		if nextToken = listRes.NextToken; nextToken == nil {
			break
		}
	}
	return nil
}

func (t *Tracer) setBoundaries(task *ecsTypes.Task) {
	d := t.option.Duration

	t.headBegin = task.CreatedAt.Add(-d)
	if task.StartedAt != nil {
		t.headEnd = task.StartedAt.Add(d)
	} else {
		t.headEnd = task.CreatedAt.Add(d)
	}
	// logs are not output before pull stopped
	if task.PullStoppedAt != nil {
		t.headEnd = task.PullStoppedAt.Add(d)
	}

	if task.StoppingAt != nil {
		t.tailBegin = task.StoppingAt.Add(-d)
	} else {
		t.tailBegin = t.now.Add(-d)
	}
	if task.StoppedAt != nil {
		t.tailEnd = task.StoppedAt.Add(d)
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

type task struct {
	Arn            string `json:"arn"`
	TaskDefinition string `json:"task_definition"`
	LastStatus     string `json:"last_status"`
	DesiredStatus  string `json:"desired_status"`
	CreatedAt      string `json:"created_at"`
	Group          string `json:"group"`
	LaunchType     string `json:"launch_type"`
}

func newTask(t *ecsTypes.Task) *task {
	return &task{
		Arn:            arnToName(*t.TaskArn),
		TaskDefinition: arnToName(*t.TaskDefinitionArn),
		LastStatus:     aws.ToString(t.LastStatus),
		DesiredStatus:  aws.ToString(t.DesiredStatus),
		CreatedAt:      t.CreatedAt.In(time.Local).Format(time.RFC3339),
		Group:          aws.ToString(t.Group),
		LaunchType:     string(t.LaunchType),
	}
}

func (t *task) String() string {
	return strings.Join([]string{
		t.Arn,
		t.TaskDefinition,
		t.LastStatus,
		t.DesiredStatus,
		t.CreatedAt,
		t.Group,
		t.LaunchType,
	}, "\t") + "\n"
}

func (t *task) JSON() string {
	b, _ := json.Marshal(t)
	return string(b) + "\n"
}
