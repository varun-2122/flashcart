.PHONY: run build test lint docker-up docker-down clean

GOPATH_BIN ?= $(shell go env GOPATH)/bin

run:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server/main.go

test:
	go test -v -race ./...

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

clean:
	rm -rf bin/
