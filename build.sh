#!/usr/bin/env bash

go fmt ./...

go build -v ./cmd/server
