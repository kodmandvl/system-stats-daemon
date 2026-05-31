BIN_DAEMON := ./bin/daemon
BIN_CLIENT := ./bin/client
DOCKER_IMG_DAEMON := system-stats-daemon:latest
GOLANGCI_VERSION := v1.63.4

.PHONY: build build_daemon build_client test lint generate integration-test

build: build_daemon build_client

build_daemon:
	go build -o $(BIN_DAEMON) ./cmd/daemon

build_client:
	go build -o $(BIN_CLIENT) ./cmd/client

test:
	go test -race -count 100 ./...

lint:
	@which golangci-lint >/dev/null || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_VERSION)
	golangci-lint run ./...

generate:
	go generate ./internal/grpc/pb/...

integration-test:
	go test -tags=integration -race -timeout 3m ./integration/...

docker-build:
	docker build -t $(DOCKER_IMG_DAEMON) -f Dockerfile .

run-daemon: build_daemon
	$(BIN_DAEMON) -port 8080 -config ./configs/daemon.json

run-client: build_client
	$(BIN_CLIENT) -config ./configs/client.json
