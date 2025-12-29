package wol

import (
	"fmt"
	"log"
	"testing"
)

func Test_WOL(t *testing.T) {
	ipFlag := "192.168.1.100"
	portFlag := 9
	macFlag := "00:11:22:33:44:55"

	// Create WOL instance
	wolConfig, err := NewWOL(ipFlag, portFlag, macFlag)
	if err != nil {
		log.Fatalf("Failed to create WOL instance: %v", err)
	}

	// Send magic packet
	fmt.Printf("Sending Wake-on-LAN magic packet to %s (IP: %s, Port: %d)...\n", macFlag, ipFlag, portFlag)

	err = wolConfig.SendMagicPacket()
	if err != nil {
		log.Fatalf("Failed to send magic packet: %v", err)
	}

	fmt.Println("Magic packet sent successfully!")
}
