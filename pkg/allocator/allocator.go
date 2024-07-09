package allocator

import (
	"hornbill/pkg/model"
	"net"
	"time"
)

type AddressSlot struct {
	Active   bool
	Expiry   time.Time
	Address  net.IP
	Identity model.Identity
}

type Allocator struct {
	Network      net.IPNet
	Cursor       net.IP
	Size         int
	AddressTable []AddressSlot
	UserMap      map[string]int
}

func NewAllocator(network net.IPNet) *Allocator {
	netSize := CalculateNetworkSizeExcludeRestricted(network)
	slotTable := make([]AddressSlot, netSize)
	ip := CloneIP(network.IP)
	for i := 0; i < netSize; i++ {
		ip = IncrementIPBound(ip, network)
		slotTable[i].Address = ip
	}

	return &Allocator{
		Network:      network,
		Cursor:       network.IP,
		Size:         netSize,
		AddressTable: slotTable,
		UserMap:      make(map[string]int),
	}
}

func (a *Allocator) Allocate(identity model.Identity) (*model.User, bool) {
	targetIndex := -1
	currentTime := time.Now()
	a.Free(identity)

	for i, v := range a.AddressTable {
		if v.Active && !v.Expiry.IsZero() && v.Expiry.Before(currentTime) {
			a.Free(v.Identity)
		}

		if !a.AddressTable[i].Active {
			targetIndex = i
			break
		}
	}

	if targetIndex > -1 {
		a.AddressTable[targetIndex].Active = true
		a.AddressTable[targetIndex].Identity = identity
		a.UserMap[identity.Id] = targetIndex

		return &model.User{
			Identity: a.AddressTable[targetIndex].Identity,
			Address:  a.AddressTable[targetIndex].Address,
		}, true
	} else {
		return nil, false
	}
}

func (a *Allocator) SetExpiry(identity model.Identity, expiry time.Time) bool {
	index, ok := a.UserMap[identity.Id]
	if !ok {
		return false
	}

	a.AddressTable[index].Expiry = expiry
	return true
}

func (a *Allocator) Get(identity model.Identity) (*model.User, bool) {
	index, ok := a.UserMap[identity.Id]
	if !ok {
		return nil, ok
	}

	slot := a.AddressTable[index]
	return &model.User{
		Identity: slot.Identity,
		Address:  slot.Address,
	}, true
}

func (a *Allocator) Free(identity model.Identity) bool {
	index, ok := a.UserMap[identity.Id]
	if !ok {
		return ok
	}

	a.AddressTable[index].Active = false
	a.AddressTable[index].Identity = model.Identity{}
	a.AddressTable[index].Expiry = time.Time{}
	delete(a.UserMap, identity.Id)
	return true
}

func (a *Allocator) Purge() {
	currentTime := time.Now()
	for _, v := range a.AddressTable {
		if v.Active && !v.Expiry.IsZero() && v.Expiry.Before(currentTime) {
			a.Free(v.Identity)
		}
	}
}

func (a *Allocator) ListUser() []model.User {
	count := len(a.UserMap)
	result := make([]model.User, 0, count)
	currentTime := time.Now()
	for _, v := range a.AddressTable {
		if !v.Active {
			continue
		}

		if !v.Expiry.IsZero() && v.Expiry.Before(currentTime) {
			a.Free(v.Identity)
			continue
		}

		result = append(result, model.User{
			Identity: v.Identity,
			Address:  v.Address,
		})
	}
	return result
}
