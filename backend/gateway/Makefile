grpc:
	protoc ../proto/auth.proto \
		--go_out=./internal/gapi/ \
		--go_opt=paths=import \
		--go-grpc_out=./internal/gapi/ \
		--go-grpc_opt=paths=import \
		--proto_path=../proto/

test:
	go test ./... -v -cover

test-cover:
	go test ./... -v -coverprofile=./test/coverage/coverage.out ./test/coverage
	go tool cover -html=./test/coverage/coverage.out -o ./test/coverage/coverage.html

.PHONY: grpc test test-cover