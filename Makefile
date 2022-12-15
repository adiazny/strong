.PHONY: tidy fmt lint test build

SHELL := /bin/bash

PROJECT_NAME="strong"

run:
	go run ./cmd/api/ -port 4001

tidy:
	go mod tidy

fmt:
	gofmt -w -s -d .

lint:
	golangci-lint run ./...

test:
	go test -cover -race ./...

build:
	@rm -f cmd/api/${PROJECT_NAME}
	@GOARCH=amd64 GOOS=linux go build -o cmd/api/${PROJECT_NAME} ./cmd/api/