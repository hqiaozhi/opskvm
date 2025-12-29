package msd

import (
	"context"
	"fmt"
)

// DisabledPlugin implements the Msd interface for disabled mode
type DisabledPlugin struct{}

// NewDisabledPlugin creates a new disabled plugin instance
func NewDisabledPlugin() *DisabledPlugin {
	return &DisabledPlugin{}
}

// GetState returns the current state of the MSD
func (p *DisabledPlugin) GetState(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"connected": false,
		"busy":      false,
		"image":     "",
		"cdrom":     false,
		"rw":        false,
		"images":    []string{},
	}, nil
}

// SetConnected sets the connected state of the MSD
func (p *DisabledPlugin) SetConnected(ctx context.Context, connected bool) error {
	return fmt.Errorf("MSD is disabled")
}

// SetParams sets parameters for the MSD
func (p *DisabledPlugin) SetParams(ctx context.Context, name string, cdrom bool, rw bool) error {
	return fmt.Errorf("MSD is disabled")
}

// ReadImage opens an image for reading
func (p *DisabledPlugin) ReadImage(ctx context.Context, name string) (MsdReader, error) {
	return nil, fmt.Errorf("MSD is disabled")
}

// WriteImage opens an image for writing
func (p *DisabledPlugin) WriteImage(ctx context.Context, name string, size int64, removeIncomplete bool) (MsdWriter, error) {
	return nil, fmt.Errorf("MSD is disabled")
}

// RemoveImage removes an image
func (p *DisabledPlugin) RemoveImage(ctx context.Context, name string) error {
	return fmt.Errorf("MSD is disabled")
}

// MakeImage creates a new image
func (p *DisabledPlugin) MakeImage(ctx context.Context, zipped bool) error {
	return fmt.Errorf("MSD is disabled")
}
