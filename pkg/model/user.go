package model

import (
	"hornbill/pkg/pb"
	"net"
	"time"
)

type Identity struct {
	Id        string
	PublicKey string
	Expiry    *time.Time
}

func IdentityFromProto(identity *pb.Identity) Identity {
	modelIdentity := Identity{
		Id:        identity.Id,
		PublicKey: identity.PublicKey,
	}
	if identity.Expiry != nil {
		milli := time.UnixMilli(*identity.Expiry)
		modelIdentity.Expiry = &milli
	}
	return modelIdentity
}

func IdentityToProto(identity Identity) *pb.Identity {
	identityPb := pb.Identity{
		Id:        identity.Id,
		PublicKey: identity.PublicKey,
	}

	if identity.Expiry != nil {
		expiryInt := identity.Expiry.UnixMilli()
		identityPb.Expiry = &expiryInt
	}

	return &identityPb
}

type User struct {
	Identity Identity
	Address  net.IP
}

func UserToProto(u *User) *pb.User {
	return &pb.User{
		Identity: IdentityToProto(u.Identity),
		Address:  u.Address.String(),
	}
}
