package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

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
	t, err := tracer.NewWithSession(ctx, sess)
	if err != nil {
		panic(err)
	}
	flag.DurationVar(&t.Duration, "duration", time.Minute, "fetch logs duration from created / before stopping")
	flag.Parse()

	args := flag.Args()
	switch len(args) {
	case 0:
		args = []string{"", ""}
	case 1:
		args = append(args, "")
	}
	if err := t.Run(args[0], args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: tracer <cluster> <task-id>")
	os.Exit(1)
}
