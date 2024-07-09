package rpcconn

import (
	"google.golang.org/grpc"
)

func NewServer() (*grpc.Server, error) {
	serverOptions := make([]grpc.ServerOption, 0, 4)

	credential, err := NewTransportCredential(ServerTransportType)
	if err != nil {
		return nil, err
	}
	if credential != nil {
		serverOptions = append(serverOptions, grpc.Creds(*credential))
	}

	return grpc.NewServer(serverOptions...), nil
}
