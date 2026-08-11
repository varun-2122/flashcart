.PHONY: run build test lint docker-up docker-down clean metrics grafana

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

# Phase 3/4: Observability Targets
metrics:
	@echo "Opening Prometheus UI at http://localhost:9090"
	@start http://localhost:9090 2>/dev/null || open http://localhost:9090

grafana:
	@echo "Opening Grafana Dashboard at http://localhost:3000 (admin / flashcart)"
	@start http://localhost:3000 2>/dev/null || open http://localhost:3000

jaeger:
	@echo "Opening Jaeger Tracing UI at http://localhost:16686"
	@start http://localhost:16686 2>/dev/null || open http://localhost:16686
