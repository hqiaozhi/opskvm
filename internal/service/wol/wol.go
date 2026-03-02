package wol

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"syscall"
)

type WOLInterface interface {
	WakeUp(broadcast_ip string, port int, mac string) error
}

// WOL represents a Wake-on-LAN configuration
type WOL struct {
	IP   string
	Port int
	MAC  string
}

// NewWOL creates a new WOL instance with validation
func NewWOL() WOLInterface {
	return &WOL{}
}

// ValidMAC validates a MAC address format
func ValidMAC(mac string) bool {
	// Regular expression for MAC address validation
	// Accepts formats like: 00:11:22:33:44:55, 00-11-22-33-44-55, 001122334455
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]?){5}([0-9A-Fa-f]{2})$`)
	return macRegex.MatchString(mac)
}

// SendMagicPacket sends a Wake-on-LAN magic packet
func (w *WOL) WakeUp(broadcast_ip string, port int, mac string) error {
	// Validate MAC address
	if !ValidMAC(mac) {
		return fmt.Errorf("invalid MAC address format: %s", mac)
	}

	// Validate IP address
	if broadcast_ip != "" && net.ParseIP(broadcast_ip) == nil {
		return fmt.Errorf("invalid IP address: %s", broadcast_ip)
	}

	// Validate port
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d", port)
	}

	// Create magic packet
	// Magic packet format: 6 bytes of FF followed by 16 repetitions of the MAC address
	macBytes, err := w.macToBytesByMac(mac)
	if err != nil {
		return err
	}

	// Create packet buffer
	packet := make([]byte, 102)

	// Fill first 6 bytes with FF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}

	// Fill the rest with 16 repetitions of the MAC address
	for i := 6; i < 102; i += 6 {
		copy(packet[i:i+6], macBytes)
	}

	// Create UDP socket
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.ParseIP(broadcast_ip),
		Port: port,
	})
	if err != nil {
		return fmt.Errorf("failed to create UDP socket: %w", err)
	}
	defer socket.Close()

	// Set broadcast flag
	fd, err := socket.File()
	if err != nil {
		return fmt.Errorf("failed to get socket file descriptor: %w", err)
	}
	defer fd.Close()

	if err := syscall.SetsockoptInt(int(fd.Fd()), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1); err != nil {
		return fmt.Errorf("failed to set broadcast option: %w", err)
	}

	// Send the magic packet
	_, err = socket.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send magic packet: %w", err)
	}

	return nil
}

// macToBytesByMac converts a MAC address string to bytes
func (w *WOL) macToBytesByMac(mac string) ([]byte, error) {
	// Remove separators
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")

	// Check length
	if len(mac) != 12 {
		return nil, fmt.Errorf("invalid MAC address length: %s", mac)
	}

	// Convert to bytes
	bytes := make([]byte, 6)
	for i := 0; i < 6; i++ {
		var b byte
		_, err := fmt.Sscanf(mac[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return nil, fmt.Errorf("failed to parse MAC address: %w", err)
		}
		bytes[i] = b
	}

	return bytes, nil
}
