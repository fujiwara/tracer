package tracer_test

import (
	"testing"

	"github.com/fujiwara/tracer"
)

func TestExtractClusterName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "arn:aws:ecs:ap-northeast-1:012345678901:cluster/main",
			expected: "main",
		},
		{
			input:    "main",
			expected: "main",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			actual := tracer.ExtractClusterName(c.input)
			if c.expected != actual {
				t.Errorf("expected: %s, actual: %s", c.expected, actual)
			}
		})
	}
}

func TestExtractTaskID(t *testing.T) {
	cases := []struct {
		input    string
		cluster  string
		expected string
	}{
		{
			input:    "0123456789abcdef0123456789abcdef",
			cluster:  "main",
			expected: "0123456789abcdef0123456789abcdef",
		},
		{
			input:    "arn:aws:ecs:ap-northeast-1:012345678901:task/main/0123456789abcdef0123456789abcdef",
			cluster:  "main",
			expected: "0123456789abcdef0123456789abcdef",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			actual := tracer.ExtractTaskID(c.cluster, c.input)
			if c.expected != actual {
				t.Errorf("expected: %s, actual: %s", c.expected, actual)
			}
		})
	}
}
