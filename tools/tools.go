//go:build tools
// +build tools

package tools

/*
	https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md
*/
import (
	_ "github.com/twitchtv/twirp/protoc-gen-twirp"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
