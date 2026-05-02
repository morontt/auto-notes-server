.PHONY: build
build:
	go fmt ./...
	go build -v ./cmd/server

.PHONY: generate
generate:
	go generate ./...
