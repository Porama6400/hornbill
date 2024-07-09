package allocator

import (
	"bytes"
	"github.com/mikioh/ipaddr"
	"net"
	"reflect"
	"testing"
)

func TestIncrement(t *testing.T) {
	testData := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "Increment single byte",
			input:    []byte{0},
			expected: []byte{1},
		},
		{
			name:     "Increment single byte overflow",
			input:    []byte{0xff},
			expected: []byte{0x00},
		},
		{
			name:     "Increment multi-byte",
			input:    []byte{0x00, 0x01},
			expected: []byte{0x00, 0x02},
		},
		{
			name:     "Increment multi-byte overflow",
			input:    []byte{0x00, 0xff},
			expected: []byte{0x01, 0x00},
		},
		{
			name:     "Increment multi-byte overflow",
			input:    []byte{0xff, 0xff},
			expected: []byte{0x00, 0x00},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Clone(tt.input)
			Increment(got)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got: %v, want: %v", got, tt.expected)
			}
		})
	}
}

func TestCloneIP(t *testing.T) {
	ip := net.IPv4(1, 2, 3, 4)
	clone := CloneIP(ip)
	if !reflect.DeepEqual(ip, clone) {
		t.Errorf("cloned value is not equal to original value")
	}
	IncrementIP(clone)
	if reflect.DeepEqual(ip, clone) {
		t.Errorf("clone is still referencing the original")
	}
}

func TestIncrementIP(t *testing.T) {
	testData := []struct {
		name     string
		input    net.IP
		expected net.IP
	}{
		{
			name:     "increment",
			input:    net.IPv4(192, 168, 1, 1),
			expected: net.IPv4(192, 168, 1, 2),
		},
		{
			name:     "increment overflow",
			input:    net.IPv4(192, 168, 1, 255),
			expected: net.IPv4(192, 168, 2, 0),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			got := make(net.IP, len(tt.input))
			copy(got, tt.input)
			IncrementIP(got)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got: %v, want: %v", got, tt.expected)
			}
		})
	}
}

func TestIncrementIPBound(t *testing.T) {
	testData := []struct {
		name   string
		ipNet  net.IPNet
		input  net.IP
		output net.IP
	}{
		{
			"increment skip router address",
			net.IPNet{
				IP:   net.IPv4(10, 0, 0, 0),
				Mask: net.IPMask{255, 255, 0, 0},
			},
			net.IPv4(10, 0, 0, 0),
			net.IPv4(10, 0, 0, 2),
		},
		{
			"increment overflow",
			net.IPNet{
				IP:   net.IPv4(10, 0, 0, 0),
				Mask: net.IPMask{255, 255, 0, 0},
			},
			net.IPv4(10, 0, 0, 255),
			net.IPv4(10, 0, 1, 0),
		},
		{
			"increment overflow skip broadcast and network and router address",
			net.IPNet{
				IP:   net.IPv4(10, 0, 0, 0),
				Mask: net.IPMask{255, 255, 0, 0},
			},
			net.IPv4(10, 0, 255, 254),
			net.IPv4(10, 0, 0, 2),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			got := IncrementIPBound(tt.input, tt.ipNet)
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("got: %v, want: %v", got, tt.output)
			}
		})
	}
}

func TestCalculateSize(t *testing.T) {
	testData := []struct {
		input    string
		expected int
	}{
		{
			"192.168.1.0/32",
			0,
		},
		{
			"192.168.1.0/30",
			1,
		},
		{
			"192.168.1.0/24",
			256 - 3,
		},
		{
			"192.168.1.0/16",
			(256 * 256) - 3,
		},
	}

	for _, tt := range testData {
		t.Run(tt.input, func(t *testing.T) {
			parse, err := ipaddr.Parse(tt.input)
			if err != nil {
				t.Error(err)
			}

			pos := parse.Pos()
			if pos == nil {
				t.Errorf("got nil pointer")
			}

			got := CalculateNetworkSizeExcludeRestricted(pos.Prefix.IPNet)
			if got != tt.expected {
				t.Errorf("got: %v, want: %v", got, tt.expected)
			}
		})
	}
}
