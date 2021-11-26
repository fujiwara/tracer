package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/fujiwara/tracer"
)

func main() {
	ctx := context.Background()
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		panic(err)
	}
	t, err := tracer.New(ctx, sess)
	if err != nil {
		panic(err)
	}
	if err := t.Run(os.Args[1], os.Args[2]); err != nil {
		panic(err)
	}
}
