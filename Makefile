BINARY ?= bin/clustercost-agent-k8s

.PHONY: build run lint test tidy

build:
	@mkdir -p $(dir $(BINARY))
	GO111MODULE=on go build -o $(BINARY) ./cmd/agent

run:
	go run ./cmd/agent

lint:
	golangci-lint run ./...

test:
	go test ./...

tidy:
	go mod tidy
