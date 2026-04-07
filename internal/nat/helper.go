package nat

import (
	"net"
)

// ExtractPublicIP extracts the public IP address from a remote address string
// Format: "203.0.113.1:12345" -> "203.0.113.1"
func ExtractPublicIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// If parsing fails, return the original address
		return remoteAddr
	}
	return host
}

// ExtractPublicPort extracts the public port from a remote address string
// Format: "203.0.113.1:12345" -> 12345
func ExtractPublicPort(remoteAddr string) int {
	addr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		return 0
	}
	return addr.Port
}

// IsPrivateIP checks if an IP address is private (RFC1918)
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IP ranges
	private := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range private {
		_, subnet, _ := net.ParseCIDR(cidr)
		if subnet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// EnrichPeerInfo adds public IP information to peer info
func EnrichPeerInfo(localIP string, localPort int, remoteAddr string) (string, int) {
	publicIP := ExtractPublicIP(remoteAddr)
	publicPort := ExtractPublicPort(remoteAddr)
	return publicIP, publicPort
}
