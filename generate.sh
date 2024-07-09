protoc daemon.proto --go_out=. --go-grpc_out=.
swag init -d cmd/apiserver,pkg/apiserver --pd
