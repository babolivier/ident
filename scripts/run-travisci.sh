#!/bin/bash

if [ $SUITE == "build" ]; then
    go build
fi

if [ $SUITE == "test" ]; then
    go test -v ./...
    go test -v -race ./...
fi

if [ $SUITE == "codecov" ]; then
    go test -v -coverprofile=coverage.txt -covermode=atomic ./...
fi