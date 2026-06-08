FROM golang:1.23 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /daemon ./cmd/daemon

FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y --no-install-recommends sysstat iproute2 procps && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /daemon /usr/local/bin/daemon
COPY configs/daemon.json /etc/stats/daemon.json

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/daemon", "-port", "8080", "-config", "/etc/stats/daemon.json"]
