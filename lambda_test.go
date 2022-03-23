package tracer

import (
	"testing"

	"github.com/stretchr/testify/require"
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
			actual := extractClusterName(c.input)
			require.Equal(t, c.expected, actual)
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
			actual := extractTaskID(c.cluster, c.input)
			require.Equal(t, c.expected, actual)
		})
	}
}
