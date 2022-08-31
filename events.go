package tracer

import (
	"bytes"
	"encoding/json"
)

type ECSTaskEvent struct {
	Detail ECSTaskEventDetail `json:"detail"`
}

type ECSTaskEventDetail struct {
	DesiredStatus string    `json:"desiredStatus"`
	LastStatus    string    `json:"lastStatus"`
	StopCode      string    `json:"stopCode"`
	StoppedReason string    `json:"stoppedReason"`
	TaskArn       string    `json:"taskArn"`
	ClusterArn    string    `json:"clusterArn"`
}

func (e *ECSTaskEvent) String() string {
	b := bytes.Buffer{}
	json.NewEncoder(&b).Encode(e)
	return b.String()
}
