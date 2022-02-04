package tracer

import (
	"context"
	"fmt"
)

func (t *Tracer) LambdaHandler(ctx context.Context, event *ECSTaskEvent) error {
	fmt.Println(event.String())
	lastStatus := event.Detail.LastStatus
	if lastStatus != "STOPPED" {
		return nil
	}
	return t.Run(event.Detail.ClusterArn, event.Detail.TaskArn)
}
