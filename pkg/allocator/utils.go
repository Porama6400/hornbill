package allocator

import (
	"bytes"
	"math"
	"net"
)

func Increment(data []byte) {
	dataLength := len(data)
	for i := dataLength - 1; i >= 0; i-- {
		if data[i] < 0xff {
			data[i]++
			break
		} else {
			data[i] = 0
		}
	}
}

func CloneIP(ip net.IP) net.IP {
	clone := make(net.IP, len(ip))
	copy(clone, ip)
	return clone
}

func IncrementIP(ip net.IP) {
	Increment(ip)
}

func GetRouterAddress(network net.IPNet) net.IP {
	ip := CloneIP(network.IP)
	IncrementIP(ip)
	return ip
}

func GetFirstAddress(network net.IPNet) net.IP {
	ip := GetRouterAddress(network)
	IncrementIP(ip)
	return ip
}

func IncrementIPBound(ip net.IP, network net.IPNet) net.IP {
	newIp := CloneIP(ip)
	IncrementIP(newIp)
	nextIp := CloneIP(newIp)
	IncrementIP(nextIp)

	if !network.Contains(newIp) || !network.Contains(nextIp) { // make sure to skip broadcast address
		newIp = CloneIP(network.IP)
		IncrementIP(newIp) // network addr
		IncrementIP(newIp) // router addr
		return newIp
	} else if bytes.Equal(newIp, network.IP) || bytes.Equal(newIp, GetRouterAddress(network)) {
		return GetFirstAddress(network)
	} else {
		return newIp
	}
}

// CalculateNetworkSizeExcludeRestricted calculate number of clients that can fit into the network
// excluding network address, broadcast address and router address
func CalculateNetworkSizeExcludeRestricted(network net.IPNet) int {
	ones, bits := network.Mask.Size()
	networkSize := math.Pow(2, float64(bits-ones)) - 3
	if networkSize < 0 {
		return 0
	} else {
		return int(networkSize)
	}
}
