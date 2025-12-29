package msd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Storage represents a storage manager for MSD images
type Storage struct {
	path string
	zip  bool
}

// NewStorage creates a new Storage instance
func NewStorage(path string, zip bool) (*Storage, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &Storage{
		path: path,
		zip:  zip,
	}, nil
}

// GetImages returns the list of available images
func (s *Storage) GetImages() ([]string, error) {
	var images []string

	entries, err := os.ReadDir(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".img") || strings.HasSuffix(name, ".iso") {
			images = append(images, strings.TrimSuffix(name, filepath.Ext(name)))
		}
	}

	return images, nil
}

// GetImagePath returns the full path to an image
func (s *Storage) GetImagePath(name string) (string, error) {
	images, err := s.GetImages()
	if err != nil {
		return "", err
	}

	for _, img := range images {
		if img == name {
			// Check for both img and iso extensions
			for _, ext := range []string{".img", ".iso"} {
				path := filepath.Join(s.path, name+ext)
				if _, err := os.Stat(path); err == nil {
					return path, nil
				}
			}
		}
	}

	return "", ErrMsdUnknownImage
}

// HasImage checks if an image exists
func (s *Storage) HasImage(name string) (bool, error) {
	_, err := s.GetImagePath(name)
	if err != nil {
		if errors.Is(err, ErrMsdUnknownImage) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetImageInfo returns information about an image
func (s *Storage) GetImageInfo(name string) (map[string]interface{}, error) {
	path, err := s.GetImagePath(name)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}

	return map[string]interface{}{
		"name":  name,
		"size":  info.Size(),
		"mtime": info.ModTime(),
		"path":  path,
	}, nil
}

// CreateImage creates a new image file
func (s *Storage) CreateImage(name string, size int64) (string, error) {
	if exists, err := s.HasImage(name); err != nil {
		return "", err
	} else if exists {
		return "", fmt.Errorf("image already exists: %s", name)
	}

	path := filepath.Join(s.path, name+".img")

	// Create the file with the specified size
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	if err := file.Truncate(size); err != nil {
		os.Remove(path) // Clean up
		return "", fmt.Errorf("failed to set image size: %w", err)
	}

	return path, nil
}

// RemoveImage removes an image file
func (s *Storage) RemoveImage(name string) error {
	path, err := s.GetImagePath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove image: %w", err)
	}

	return nil
}

// OpenImageForReading opens an image for reading
func (s *Storage) OpenImageForReading(name string) (*os.File, error) {
	path, err := s.GetImagePath(name)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image for reading: %w", err)
	}

	return file, nil
}

// OpenImageForWriting opens an image for writing
func (s *Storage) OpenImageForWriting(name string, size int64, removeIncomplete bool) (*os.File, error) {
	path, err := s.GetImagePath(name)
	if err != nil {
		// If image doesn't exist, create it
		if errors.Is(err, ErrMsdUnknownImage) {
			path, err = s.CreateImage(name, size)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// Check if image exists and is writeable
		if err := checkImageWriteable(path); err != nil {
			return nil, err
		}

		// Truncate or expand to the specified size
		if err := os.Truncate(path, size); err != nil {
			return nil, fmt.Errorf("failed to resize image: %w", err)
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		if removeIncomplete {
			os.Remove(path)
		}
		return nil, fmt.Errorf("failed to open image for writing: %w", err)
	}

	return file, nil
}

// checkImageWriteable checks if an image file is writeable
func checkImageWriteable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Mode()&0200 == 0 { // Check user write permission
		return ErrMsdImageIsReadOnly
	}

	return nil
}
