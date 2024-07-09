package daemon

import (
	"context"
	"hornbill/pkg/allocator"
	"hornbill/pkg/model"
	"hornbill/pkg/pb"
)

type Server struct {
	pb.UnimplementedDaemonServer
	Allocator *allocator.Allocator
	WireGuard *WireGuard
}

func (s *Server) Ping(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *Server) Tick(ctx context.Context, _ *pb.Empty) (*pb.Result, error) {
	_, err := s.WireGuard.Configure(s.Allocator.ListUser())
	if err != nil {
		return nil, err
	}

	return &pb.Result{}, nil
}

func (s *Server) Add(ctx context.Context, identityPb *pb.Identity) (*pb.ResultAdd, error) {
	identity := model.IdentityFromProto(identityPb)
	result, ok := s.Allocator.Allocate(identity)
	if !ok {
		return &pb.ResultAdd{
			Ok: false,
		}, nil
	}

	if identity.Expiry != nil {
		s.Allocator.SetExpiry(identity, *identity.Expiry)
	}

	ok, err := s.WireGuard.Configure(s.Allocator.ListUser())
	if err != nil {
		s.Allocator.Free(identity)
		return &pb.ResultAdd{
			Ok:      false,
			Message: "WireGuard reconfiguration failed",
		}, err
	}

	if !ok {
		s.Allocator.Free(identity)
		return &pb.ResultAdd{
			Ok:      false,
			Message: "WireGuard reconfiguration failed",
		}, nil
	}

	userProto := model.UserToProto(result)
	return &pb.ResultAdd{
		Ok:   true,
		User: userProto,
		ServerInfo: &pb.ServerInfo{
			PublicKey:      s.WireGuard.Config.PublicKey.String(),
			PublicAddress:  s.WireGuard.Config.PublicAddress,
			AllowedAddress: s.WireGuard.Config.AllowedAddress,
		},
	}, nil
}
func (s *Server) Remove(ctx context.Context, identityPb *pb.Identity) (*pb.Result, error) {
	identity := model.IdentityFromProto(identityPb)
	ok := s.Allocator.Free(identity)
	if !ok {
		return &pb.Result{Ok: false, Message: "free failed"}, nil
	}

	_, err := s.WireGuard.Configure(s.Allocator.ListUser())
	if err != nil {
		return &pb.Result{
			Ok: false,
		}, err
	}

	return &pb.Result{Ok: true}, nil
}
func (s *Server) List(ctx context.Context, empty *pb.Empty) (*pb.UserList, error) {
	users := s.Allocator.ListUser()
	resultUsers := make([]*pb.User, 0)
	for _, user := range users {
		resultUsers = append(resultUsers, model.UserToProto(&user))
	}
	result := &pb.UserList{Users: resultUsers}
	return result, nil
}
