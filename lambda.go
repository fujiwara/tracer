package tracer

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

func (t *Tracer) LambdaHandlerFunc(opt *RunOption) func(ctx context.Context, event *ECSTaskEvent) error {
	return func(ctx context.Context, event *ECSTaskEvent) error {
		fmt.Println(event.String())
		lastStatus := event.Detail.LastStatus
		if lastStatus != "STOPPED" {
			return nil
		}
		cluster := extractClusterName(event.Detail.ClusterArn)
		return t.Run(ctx, cluster, extractTaskID(cluster, event.Detail.TaskArn), opt)
	}
}

func extractClusterName(clusterArn string) string {
	parsed, err := arn.Parse(clusterArn)
	if err != nil {
		return clusterArn
	}
	prefix := "cluster/"
	if parsed.Service == "ecs" && strings.HasPrefix(parsed.Resource, prefix) {
		return strings.TrimPrefix(parsed.Resource, prefix)
	}
	return clusterArn
}

func extractTaskID(cluster, taskArn string) string {
	parsed, err := arn.Parse(taskArn)
	if err != nil {
		return taskArn
	}
	prefix := "task/" + cluster + "/"
	if parsed.Service == "ecs" && strings.HasPrefix(parsed.Resource, prefix) {
		return strings.TrimPrefix(parsed.Resource, prefix)
	}
	return taskArn
}
