package net

import "net"

// Subnet применяется для проверки вхождения ip в подсеть
type Subnet struct {
	subnet *net.IPNet
}

// NewSubnet конструктор для создания Subnet
func NewSubnet(subnet string) (*Subnet, error) {
	_, subnetIP, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	return &Subnet{subnetIP}, nil
}

// Проверка входения ip в подсеть
func (s *Subnet) Contains(ip string) bool {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	return s.subnet.Contains(ipAddr)
}
