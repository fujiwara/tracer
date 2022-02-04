bootstrap:
	GOOS=linux GOARCH=amd64 go build -o bootstrap ../cmd/tracer/main.go

clean:
	rm -f bootstrap

deploy: bootstrap
	lambroll deploy --function function.jsonnet
