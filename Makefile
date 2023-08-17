protos:
	protoc --go_out=. --go-grpc_out=. proto/filebank.proto proto/signed.proto proto/storage.proto