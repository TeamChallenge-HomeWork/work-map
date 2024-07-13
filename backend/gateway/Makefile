grpc:
	protoc ../proto/auth.proto \
		--go_out=./internal/gapi/ \
		--go_opt=paths=import \
		--go-grpc_out=./internal/gapi/ \
		--go-grpc_opt=paths=import \
		--proto_path=../proto/