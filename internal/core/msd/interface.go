package msd

import (
	"context"
)

// MsdError represents the base MSD error type
type MsdError struct {
	message string
}

func (e *MsdError) Error() string {
	return e.message
}

// NewMsdError creates a new MsdError
func NewMsdError(message string) *MsdError {
	return &MsdError{message: message}
}

// MsdIsBusyError indicates that MSD is busy
var (
	ErrMsdIsBusy          = NewMsdError("MSD is busy")
	ErrMsdIsNotConnected  = NewMsdError("MSD is not connected")
	ErrMsdIsConnected     = NewMsdError("MSD is already connected")
	ErrMsdUnknownImage    = NewMsdError("MSD image is unknown")
	ErrMsdInvalidImage    = NewMsdError("MSD image is invalid")
	ErrMsdImageNotLoaded  = NewMsdError("MSD image is not loaded")
	ErrMsdImageIsDirty    = NewMsdError("MSD image is dirty")
	ErrMsdImageIsReadOnly = NewMsdError("MSD image is read-only")
	ErrMsdIsReadOnly      = NewMsdError("MSD is read-only")
	ErrMsdNotMounted      = NewMsdError("MSD is not mounted")
	ErrMsdOffline         = NewMsdError("MSD is offline")
	ErrMsdDisconnected    = NewMsdError("MSD is disconnected")
	ErrMsdImageNotSelected = NewMsdError("MSD image is not selected")
	ErrMsdImageExists     = NewMsdError("MSD image already exists")
	ErrMsdOperationError  = NewMsdError("MSD operation error")
	ErrMsdDriveLocked     = NewMsdError("MSD drive is locked on IO operation")
)

// MsdReader defines the interface for reading MSD images
type MsdReader interface {
	// Read reads up to len(p) bytes from the image.
	// It returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
	Read(p []byte) (n int, err error)

	// Close closes the reader and releases any resources.
	Close() error
}

// MsdWriter defines the interface for writing MSD images
type MsdWriter interface {
	// Write writes len(p) bytes from p to the image.
	// It returns the number of bytes written (0 <= n <= len(p)) and any error encountered.
	Write(p []byte) (n int, err error)

	// Commit commits the written data to the image.
	Commit() error

	// Close closes the writer and releases any resources.
	Close() error
}

// Msd defines the main interface for MSD operations
type Msd interface {
	// GetState returns the current state of the MSD
	GetState(ctx context.Context) (map[string]interface{}, error)

	// SetConnected sets the connected state of the MSD
	SetConnected(ctx context.Context, connected bool) error

	// SetParams sets parameters for the MSD
	SetParams(ctx context.Context, name string, cdrom bool, rw bool) error

	// ReadImage opens an image for reading
	ReadImage(ctx context.Context, name string) (MsdReader, error)

	// WriteImage opens an image for writing
	WriteImage(ctx context.Context, name string, size int64, removeIncomplete bool) (MsdWriter, error)

	// RemoveImage removes an image
	RemoveImage(ctx context.Context, name string) error

	// MakeImage creates a new image
	MakeImage(ctx context.Context, zipped bool) error
}
