package dhcp

import (
	"encoding/binary"
	"net"
)

func ipToU32(ip net.IP) uint32 {
	ip4 := ip.To4()
	return binary.BigEndian.Uint32(ip4)
}

func u32ToIP(u uint32) net.IP {
	b := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(b, u)
	return net.IPv4(b[0], b[1], b[2], b[3])
}

type uIPs []uint32

func (u uIPs) Len() int {
	return len(u)
}
func (u uIPs) Less(i, j int) bool {
	return u[i] < u[j]
}
func (u uIPs) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
