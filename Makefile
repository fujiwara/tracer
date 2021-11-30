tracer: *.go *.go cmd/tracer/*
	go build -o $@ cmd/tracer/main.go

install:
	go install github.com/fujiwara/tracer/cmd/tracer
