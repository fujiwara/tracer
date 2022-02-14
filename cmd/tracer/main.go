package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/fujiwara/tracer"
)

var Version = "current"

func init() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(w, "tracer [options] [cluster] [task-id]")
		fmt.Fprintln(w, "")
		flag.PrintDefaults()
	}
}

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
	var showVersion bool
	flag.DurationVar(&t.Duration, "duration", time.Minute, "fetch logs duration from created / before stopping")
	flag.BoolVar(&showVersion, "version", false, "show the version")
	flag.VisitAll(envToFlag)
	flag.Parse()

	if showVersion {
		fmt.Println("tracer", Version)
		return
	}

	if onLambda() {
		lambda.Start(t.LambdaHandler)
		return
	}
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

func onLambda() bool {
	return strings.HasPrefix(os.Getenv("AWS_EXECUTE_ENV"), "AWS_Lambda") ||
		os.Getenv("AWS_LAMBDA_RUNTIME_API") != ""
}

func envToFlag(f *flag.Flag) {
	name := strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
	if s, ok := os.LookupEnv("TRACER_" + name); ok {
		f.Value.Set(s)
	}
}
