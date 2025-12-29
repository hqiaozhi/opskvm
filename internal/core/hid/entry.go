package hid

import (
	"opskvm/internal/core/hid/ch9329"
	"opskvm/internal/core/hid/otg"
)

type HIDEntry struct {
	ch9329.HIDDevice
	otg.GadgetInterface
}
