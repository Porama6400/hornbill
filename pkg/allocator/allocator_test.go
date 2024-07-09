package allocator

import (
	"hornbill/pkg/model"
	"math/rand/v2"
	"net"
	"strconv"
	"testing"
	"time"
)

func MakeAllocator() *Allocator {
	return NewAllocator(net.IPNet{
		IP:   net.IPv4(192, 168, 1, 0),
		Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0b11111000),
	})
}

func TestAllocator(t *testing.T) {
	allocator := MakeAllocator()
	mockUsers := make([]model.Identity, allocator.Size)
	for i := range mockUsers {
		mockUsers[i].Id = strconv.Itoa(int(rand.Int32()))
	}
	t.Run("get - empty allocator", func(t *testing.T) {
		for i := 0; i < allocator.Size; i++ {
			result, ok := allocator.Get(mockUsers[i])
			if ok || result != nil {
				t.Errorf("get returned a result while allocator is empty")
			}
		}
	})
	t.Run("allocate - normal", func(t *testing.T) {
		for i := 0; i < allocator.Size; i++ {
			_, ok := allocator.Allocate(mockUsers[i])
			if !ok {
				t.Errorf("failed to allocate %+v", mockUsers[i])
			}
		}
	})
	t.Run("allocate - full allocator", func(t *testing.T) {
		_, ok := allocator.Allocate(model.Identity{
			Id:        "fail",
			PublicKey: "fail",
		})
		if ok {
			t.Errorf("allocate success when allocator is full")
		}
		if len(allocator.ListUser()) != allocator.Size {
			t.Errorf("ListUser is not the same as allocator size")
		}
	})
	t.Run("get - normal", func(t *testing.T) {
		for i := 0; i < allocator.Size; i++ {
			_, ok := allocator.Get(mockUsers[i])
			if !ok {
				t.Errorf("failed to get on a full allocator")
			}
		}
	})
	t.Run("free - normal", func(t *testing.T) {
		for i := 0; i < allocator.Size; i++ {
			ok := allocator.Free(mockUsers[i])
			if !ok {
				t.Errorf("failed to free %+v", mockUsers[i])
			}
		}
	})
	t.Run("free - empty allocator", func(t *testing.T) {
		for i := 0; i < allocator.Size; i++ {
			ok := allocator.Free(mockUsers[i])
			if ok {
				t.Errorf("able to free %+v", mockUsers[i])
			}
		}
	})
	t.Run("verify empty", func(t *testing.T) {
		for i, v := range allocator.AddressTable {
			if v.Active == true {
				t.Errorf("slot %d is not empty", i)
			}
		}
		if len(allocator.UserMap) != 0 {
			t.Errorf("UserMap is not empty")
		}
		if len(allocator.ListUser()) != 0 {
			t.Errorf("ListUser is not empty")
		}
	})
}

func TestAllocatorPurge(t *testing.T) {
	allocator := MakeAllocator()
	mockUsers := make([]model.Identity, allocator.Size)
	for i := range mockUsers {
		mockUsers[i].Id = strconv.Itoa(int(rand.Int32()))
		_, ok := allocator.Allocate(mockUsers[i])
		if !ok {
			t.Errorf("allocation failed")
		}
	}
	if len(allocator.ListUser()) != allocator.Size {
		t.Errorf("allocation failed")
	}

	t.Run("purge", func(t *testing.T) {
		allocator.AddressTable[0].Expiry = time.Now().Add(-1 * time.Hour)
		allocator.Purge()
		if len(allocator.ListUser()) != allocator.Size-1 {
			t.Errorf("ListUser return expired result")
		}
	})

	t.Run("purge via ListUser", func(t *testing.T) {
		for i := range mockUsers {
			allocator.AddressTable[i].Expiry = time.Now().Add(-1 * time.Hour)
		}
		if len(allocator.ListUser()) != 0 {
			t.Errorf("ListUser return expired result")
		}
	})
}

func TestAllocatorPurgeViaAllocate(t *testing.T) {
	allocator := MakeAllocator()
	mockUsers := make([]model.Identity, allocator.Size)
	for i := range mockUsers {
		mockUsers[i].Id = strconv.Itoa(int(rand.Int32()))
		_, ok := allocator.Allocate(mockUsers[i])
		if !ok {
			t.Errorf("allocation failed")
		}
	}
	if len(allocator.ListUser()) != allocator.Size {
		t.Errorf("allocation failed")
	}

	t.Run("test allocate with expired entry", func(t *testing.T) {
		allocator.AddressTable[0].Expiry = time.Now().Add(-1 * time.Hour)
		_, ok := allocator.Allocate(model.Identity{
			Id:        "test",
			PublicKey: "test",
		})
		if !ok {
			t.Errorf("failed to allocate when there's expired slot")
		}
	})
}
