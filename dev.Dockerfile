# syntax=docker/dockerfile:1

# Build a golang image based on https://docs.docker.com/language/golang/build-images

FROM golang:1.18 AS build

# Build Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./cmd/server/main.go ./cmd/server/main.go
COPY ./internal/ ./internal/

# Compile the application with the optimizations turned off
# This is important for the debugger to correctly work with the binary
RUN go build -gcflags "all=-N -l" -o ./server ./cmd/server/main.go

# Build the server image

FROM debian:buster

WORKDIR /root/

COPY --from=0 /app/server ./
COPY ./config/ ./config/
COPY ./migrations/ ./migrations/

EXPOSE 8080 40000

WORKDIR /
COPY --from=build /go/bin/dlv /

WORKDIR /root/

CMD ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./server"]
