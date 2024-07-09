package rpcconn

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(address string) (*grpc.ClientConn, error) {
	dialOptions := make([]grpc.DialOption, 0, 4)

	credential, err := NewTransportCredential(ClientTransportType)
	if err != nil {
		return nil, err
	}
	if credential != nil {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(*credential))
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return grpc.NewClient(address, dialOptions...)
}
