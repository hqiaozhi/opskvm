//go:build integration

package usbgadget

import (
	"log"
	"os"
	"testing"
)

// TestRelativeMouseOperation tests relative mouse operations
func TestRelativeMouseOperation(t *testing.T) {
	// Check if running as root
	if os.Geteuid() != 0 {
		t.Skip("This test requires root privileges")
	}

	log.Println("=== Testing Relative Mouse Mode ===")

	// Create relative mouse mode device instance
	relGadget := NewRelative("", "rel-mouse-test-gadget")
	if relGadget == nil {
		t.Fatal("Failed to create relative mouse gadget")
	}

	// Cleanup resources when test ends
	defer func() {
		if err := relGadget.Cleanup(); err != nil {
			log.Printf("Cleanup warning: %v", err)
		}
	}()

	// Initialize device
	if err := relGadget.Setup(); err != nil {
		t.Fatalf("Failed to initialize relative mouse device: %v", err)
	}

	// Test relative mouse movement
	t.Log("Testing relative mouse movement...")
	if err := relGadget.SendMouseRelative(MouseButtonNone, 10, 5); err != nil {
		t.Errorf("Failed to send relative mouse movement: %v", err)
	}

	// Test mouse left click
	t.Log("Testing mouse left click...")
	if err := relGadget.SendMouseRelative(MouseButtonLeft, 0, 0); err != nil {
		t.Errorf("Failed to send mouse left click: %v", err)
	}

	// Test keyboard operation
	t.Log("Testing keyboard operation...")
	if err := relGadget.SendKeyboard(KeyboardModifierCtrl, KeyA); err != nil {
		t.Errorf("Failed to send keyboard Ctrl+A: %v", err)
	}

	// Cleanup explicitly before test ends
	if err := relGadget.Cleanup(); err != nil {
		t.Errorf("Failed to cleanup relative mouse device: %v", err)
	}
}

// TestAbsoluteMouseOperation tests absolute mouse operations
func TestAbsoluteMouseOperation(t *testing.T) {
	// Check if running as root
	if os.Geteuid() != 0 {
		t.Skip("This test requires root privileges")
	}

	log.Println("\n=== Testing Absolute Mouse Mode ===")

	// Create absolute mouse mode device instance (keyboard+mouse only, no ISO)
	absGadget := NewAbsolute("", "abs-mouse-test-gadget")
	if absGadget == nil {
		t.Fatal("Failed to create absolute mouse gadget")
	}

	// Cleanup resources when test ends
	defer func() {
		if err := absGadget.Cleanup(); err != nil {
			log.Printf("Cleanup warning: %v", err)
		}
	}()

	// Initialize device
	if err := absGadget.Setup(); err != nil {
		t.Fatalf("Failed to initialize absolute mouse device: %v", err)
	}

	// Test absolute mouse movement to center
	t.Log("Testing absolute mouse movement...")
	centerX := 1920 / 2
	centerY := 1080 / 2
	if err := absGadget.SendMouseAbsolute(MouseButtonNone, centerX, centerY); err != nil {
		t.Errorf("Failed to send absolute mouse movement: %v", err)
	}

	// Test mouse left click at center
	t.Log("Testing absolute mouse left click...")
	if err := absGadget.SendMouseAbsolute(MouseButtonLeft, centerX, centerY); err != nil {
		t.Errorf("Failed to send absolute mouse left click: %v", err)
	}

	// Test keyboard operation
	t.Log("Testing keyboard operation...")
	if err := absGadget.SendKeyboard(KeyboardModifierNone, KeyEnter); err != nil {
		t.Errorf("Failed to send keyboard Enter: %v", err)
	}

	// Cleanup explicitly before test ends
	if err := absGadget.Cleanup(); err != nil {
		t.Errorf("Failed to cleanup absolute mouse device: %v", err)
	}
}
