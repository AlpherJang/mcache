proto:
	protoc --go_out=plugins=grpc:. --go-json_out=emit_defaults:. ./pkg/proto/data.proto
