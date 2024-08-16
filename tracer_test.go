package tracer_test

import (
	"strings"
	"testing"
	"time"

	"github.com/fujiwara/tracer"
)

func ptr[T any](v T) *T {
	return &v
}

var (
	testEvents = []tracer.TimeLineEvent{
		{
			Timestamp: ptr(time.Date(2021, 1, 2, 3, 4, 5, 123_999_000, time.UTC)),
			Message:   "test message 1",
			Source:    "test_source 1",
		},
		{
			Timestamp: ptr(time.Date(2021, 1, 2, 3, 4, 5, 123_999_999, time.UTC)),
			Message:   "test message 5",
			Source:    "test_source 5",
		},
		{
			Timestamp: ptr(time.Date(2021, 1, 2, 3, 4, 6, 123_999_000, time.UTC)),
			Message:   "test message 2",
			Source:    "test_source 2",
		},
		{
			// same timestamp to test sort stable
			Timestamp: ptr(time.Date(2021, 1, 2, 3, 4, 5, 123_999_000, time.UTC)),
			Message:   "test message 3",
			Source:    "test_source 3",
		},
		{
			// duplicate event
			Timestamp: ptr(time.Date(2021, 1, 2, 3, 4, 5, 123_999_000, time.UTC)),
			Message:   "test message 3",
			Source:    "test_source 3",
		},
	}
	expectedOutput = `2021-01-02T03:04:05.123Z	test_source 1	test message 1
2021-01-02T03:04:05.123Z	test_source 3	test message 3
2021-01-02T03:04:05.123Z	test_source 5	test message 5
2021-01-02T03:04:06.123Z	test_source 2	test message 2
`
)

func TestTimeLineEvent(t *testing.T) {
	t.Setenv("TZ", "UTC")
	now := time.Date(2021, 1, 2, 3, 4, 5, 123_999_000, time.UTC)
	ev := tracer.TimeLineEvent{
		Timestamp: &now,
		Message:   "test message",
		Source:    "test_source",
	}
	if ev.String() != "2021-01-02T03:04:05.123Z\ttest_source\ttest message\n" {
		t.Errorf("unexpected string: %s", ev.String())
	}
}

func TestTimeLine(t *testing.T) {
	t.Setenv("TZ", "UTC")
	tl := tracer.NewTimeline()
	for _, ev := range testEvents {
		ev := ev
		tl.Add(&ev)
	}
	b := new(strings.Builder)
	tl.Print(b)
	if b.String() != expectedOutput {
		t.Errorf("unexpected output: %s", b.String())
	}
}
