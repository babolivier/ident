#!/bin/bash

if [ $SUITE == "build" ]; then
    GOOS=linux go build
    GOOS=darwin go build
    GOOS=freebsd go build
    GOOS=windows go build
fi

if [ $SUITE == "test" ]; then
    go test -v ./...
    go test -v -race ./...
fi

if [ $SUITE == "codecov" ]; then
    go test -v -coverprofile=coverage.txt -covermode=atomic ./...
fi