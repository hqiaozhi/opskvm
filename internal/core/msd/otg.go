package msd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// OTGPlugin implements the Msd interface for OTG mode
type OTGPlugin struct {
	storage *Storage
	drive   *Drive
	busy    bool
	mu      sync.Mutex
	image   string
}

// OTGPluginConfig contains configuration for the OTG plugin
type OTGPluginConfig struct {
	StoragePath string
	DrivePath   string
	ZipEnabled  bool
}

// NewOTGPlugin creates a new OTG plugin instance
func NewOTGPlugin(config OTGPluginConfig) (*OTGPlugin, error) {
	storage, err := NewStorage(config.StoragePath, config.ZipEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	drive := NewDrive(config.DrivePath, 0, 0)

	return &OTGPlugin{
		storage: storage,
		drive:   drive,
		busy:    false,
		image:   "",
	}, nil
}

// GetState returns the current state of the MSD
func (p *OTGPlugin) GetState(ctx context.Context) (map[string]interface{}, error) {
	p.mu.Lock()
	busy := p.busy
	image := p.image
	p.mu.Unlock()

	connected, err := p.drive.IsConnected()
	if err != nil {
		return nil, err
	}

	cdrom, err := p.drive.GetCdrom()
	if err != nil {
		return nil, err
	}

	rw, err := p.drive.GetRw()
	if err != nil {
		return nil, err
	}

	state := map[string]interface{}{
		"connected": connected,
		"busy":      busy,
		"image":     image,
		"cdrom":     cdrom,
		"rw":        rw,
		"images":    []string{},
	}

	// Get available images
	images, err := p.storage.GetImages()
	if err != nil {
		return nil, err
	}
	state["images"] = images

	// Get current image info if an image is loaded
	if image != "" {
		info, err := p.storage.GetImageInfo(image)
		if err != nil {
			return nil, err
		}
		state["image_info"] = info
	}

	return state, nil
}

// SetConnected sets the connected state of the MSD
func (p *OTGPlugin) SetConnected(ctx context.Context, connected bool) error {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return ErrMsdIsBusy
	}
	p.mu.Unlock()

	currentConnected, err := p.drive.IsConnected()
	if err != nil {
		return err
	}

	if currentConnected == connected {
		return nil
	}

	if !connected {
		// Disconnect the drive
		if err := p.drive.SetConnected(false); err != nil {
			return err
		}

		// Wait for disconnection
		if err := p.drive.WaitForDisconnect(ctx, 10*time.Second); err != nil {
			return err
		}

		// Clear the image
		if err := p.drive.ClearImagePath(); err != nil {
			return err
		}

		p.mu.Lock()
		p.image = ""
		p.mu.Unlock()

		return nil
	}

	// Connect the drive
	if err := p.drive.SetConnected(true); err != nil {
		return err
	}

	// Wait for connection
	if err := p.drive.WaitForConnect(ctx, 10*time.Second); err != nil {
		return err
	}

	return nil
}

// SetParams sets parameters for the MSD
func (p *OTGPlugin) SetParams(ctx context.Context, name string, cdrom bool, rw bool) error {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return ErrMsdIsBusy
	}
	p.mu.Unlock()

	connected, err := p.drive.IsConnected()
	if err != nil {
		return err
	}

	if connected {
		return ErrMsdIsConnected
	}

	// Set CD-ROM mode if specified
	if err := p.drive.SetCdrom(cdrom); err != nil {
		return err
	}

	// Set read-write mode if specified
	if err := p.drive.SetRw(rw); err != nil {
		return err
	}

	// Set the image if provided
	if name != "" {
		p.mu.Lock()
		defer p.mu.Unlock()

		// Check if image exists
		imageExists, err := p.storage.HasImage(name)
		if err != nil {
			return err
		}

		if !imageExists {
			return ErrMsdUnknownImage
		}

		// Get image path
		imagePath, err := p.storage.GetImagePath(name)
		if err != nil {
			return err
		}

		// Set the image
		if err := p.drive.SetImagePath(imagePath); err != nil {
			return err
		}

		p.image = name
	}

	return nil
}

// ReadImage opens an image for reading
func (p *OTGPlugin) ReadImage(ctx context.Context, name string) (MsdReader, error) {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return nil, ErrMsdIsBusy
	}

	connected, err := p.drive.IsConnected()
	if err != nil {
		p.mu.Unlock()
		return nil, err
	}

	if connected {
		p.mu.Unlock()
		return nil, ErrMsdIsConnected
	}

	p.busy = true
	p.mu.Unlock()

	// Open the image file
	file, err := p.storage.OpenImageForReading(name)
	if err != nil {
		p.mu.Lock()
		p.busy = false
		p.mu.Unlock()
		return nil, err
	}

	return &msdReader{
		file:   file,
		plugin: p,
	}, nil
}

// WriteImage opens an image for writing
func (p *OTGPlugin) WriteImage(ctx context.Context, name string, size int64, removeIncomplete bool) (MsdWriter, error) {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return nil, ErrMsdIsBusy
	}

	connected, err := p.drive.IsConnected()
	if err != nil {
		p.mu.Unlock()
		return nil, err
	}

	if connected {
		p.mu.Unlock()
		return nil, ErrMsdIsConnected
	}

	p.busy = true
	p.mu.Unlock()

	// Open the image file for writing
	file, err := p.storage.OpenImageForWriting(name, size, removeIncomplete)
	if err != nil {
		p.mu.Lock()
		p.busy = false
		p.mu.Unlock()
		return nil, err
	}

	return &msdWriter{
		file:             file,
		plugin:           p,
		removeIncomplete: removeIncomplete,
	}, nil
}

// RemoveImage removes an image
func (p *OTGPlugin) RemoveImage(ctx context.Context, name string) error {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return ErrMsdIsBusy
	}

	connected, err := p.drive.IsConnected()
	if err != nil {
		p.mu.Unlock()
		return err
	}

	if connected {
		p.mu.Unlock()
		return ErrMsdIsConnected
	}

	if p.image == name {
		// Clear the image from the drive first
		if err := p.drive.ClearImagePath(); err != nil {
			p.mu.Unlock()
			return err
		}
		p.image = ""
	}
	p.mu.Unlock()

	// Remove the image from storage
	return p.storage.RemoveImage(name)
}

// MakeImage creates a new image
func (p *OTGPlugin) MakeImage(ctx context.Context, zipped bool) error {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return ErrMsdIsBusy
	}

	connected, err := p.drive.IsConnected()
	if err != nil {
		p.mu.Unlock()
		return err
	}

	if connected {
		p.mu.Unlock()
		return ErrMsdIsConnected
	}
	p.mu.Unlock()

	// TODO: Implement image creation logic
	// For now, just return an error indicating not implemented
	return fmt.Errorf("image creation not implemented")
}

// msdReader implements the MsdReader interface
type msdReader struct {
	file   *os.File
	plugin *OTGPlugin
}

// Read reads from the image file
func (r *msdReader) Read(p []byte) (n int, err error) {
	return r.file.Read(p)
}

// Close closes the reader and releases the plugin
func (r *msdReader) Close() error {
	err := r.file.Close()

	r.plugin.mu.Lock()
	r.plugin.busy = false
	r.plugin.mu.Unlock()

	return err
}

// msdWriter implements the MsdWriter interface
type msdWriter struct {
	file             *os.File
	plugin           *OTGPlugin
	removeIncomplete bool
}

// Write writes to the image file
func (w *msdWriter) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

// Commit commits the written data
func (w *msdWriter) Commit() error {
	// For a simple file writer, committing just means flushing the data
	return w.file.Sync()
}

// Close closes the writer and releases the plugin
func (w *msdWriter) Close() error {
	var err error

	if w.removeIncomplete {
		// If we're removing incomplete images and the file was not committed,
		// delete it
		w.file.Close()
		err = os.Remove(w.file.Name())
	} else {
		err = w.file.Close()
	}

	w.plugin.mu.Lock()
	w.plugin.busy = false
	w.plugin.mu.Unlock()

	return err
}
