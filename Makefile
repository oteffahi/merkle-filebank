protos:
	protoc --go_out=. --go-grpc_out=. proto/filebank.proto proto/storage.proto