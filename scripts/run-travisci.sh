#!/bin/bash -e

if [ $SUITE == "build" ]; then
    go build -v
fi

if [ $SUITE == "test" ]; then
    go test -v ./...
    go test -v -race ./...
fi

if [ $SUITE == "codecov" ]; then
    go test -v -coverprofile=coverage.txt -covermode=atomic ./...
    bash <(curl -s https://codecov.io/bash)
fi