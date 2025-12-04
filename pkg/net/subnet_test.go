package net

import (
	"testing"
)

func TestNewSubnet(t *testing.T) {
	subnetStr := "192.168.1.0/24"
	subnet, err := NewSubnet(subnetStr)
	if err != nil {
		t.Errorf("NewSubnet returned error: %v", err)
	}
	if subnet == nil {
		t.Errorf("NewSubnet returned nil")
	}
}

func TestSubnetContains(t *testing.T) {
	subnetStr := "192.168.1.0/24"
	subnet, err := NewSubnet(subnetStr)
	if err != nil {
		t.Errorf("NewSubnet returned error: %v", err)
	}

	tests := []struct {
		ip     string
		want   bool
		name   string
	}{
		{"192.168.1.100", true, "IP в подсети"},
		{"192.168.2.100", false, "IP не в подсети"},
		{"192.168.1.255", true, "IP в подсети (максимальный адрес)"},
		{"192.168.1.0", true, "IP в подсети (минимальный адрес)"},
		{"256.1.1.1", false, "Недопустимый IP-адрес"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := subnet.Contains(tt.ip); got != tt.want {
				t.Errorf("Subnet.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubnetContainsInvalidIP(t *testing.T) {
	subnetStr := "192.168.1.0/24"
	subnet, err := NewSubnet(subnetStr)
	if err != nil {
		t.Errorf("NewSubnet returned error: %v", err)
	}

	tests := []struct {
		ip     string
		want   bool
		name   string
	}{
		{"", false, "Пустой IP-адрес"},
		{"abc", false, "Недопустимый IP-адрес"},
		{"192.168.1", false, "Недопустимый IP-адрес"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := subnet.Contains(tt.ip); got != tt.want {
				t.Errorf("Subnet.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSubnetInvalid(t *testing.T) {
	tests := []struct {
		subnet string
		want   error
		name   string
	}{
		{"", nil, "Пустая подсеть"},
		{"abc", nil, "Недопустимая подсеть"},
		{"192.168.1", nil, "Недопустимая подсеть"},
		{"192.168.1.0/33", nil, "Недопустимая подсеть"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSubnet(tt.subnet)
			if err == nil {
				t.Errorf("NewSubnet returned nil error")
			}
		})
	}
}