#!/usr/bin/env bash

go fmt ./...

cd ./cmd/server && go build -v
