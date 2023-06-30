package vm

import (
	"crypto/rand"
	"fmt"
)

func generateMACAddress() (string, error) {
	// MAC address is 6 bytes long
	mac := make([]byte, 6)

	// Read 6 random bytes
	_, err := rand.Read(mac)
	if err != nil {
		return "", err
	}

	// Set the locally administered bit, and clear the multicast bit to make it a valid MAC address
	// Locally administered bit is the second least significant bit of the first byte
	// Multicast bit is the least significant bit of the first byte
	mac[0] = (mac[0] & 0xfe) | 0x02

	// Format the MAC address
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]), nil
}
