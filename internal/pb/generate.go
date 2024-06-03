package pb

//go:generate protoc --proto_path=./../../proto --go_out=. --go_opt=paths=source_relative --twirp_out=. --twirp_opt=paths=source_relative auth.proto
func init() {}
