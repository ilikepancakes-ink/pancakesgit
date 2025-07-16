package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"pancakesgit/internal/config"
)

// Service provides storage functionality
type Service struct {
	config config.StorageConfig
	basePath string
}

// New creates a new storage service
func New(cfg config.StorageConfig) *Service {
	return &Service{
		config: cfg,
		basePath: cfg.Path,
	}
}

// Store stores data at the given path
func (s *Service) Store(path string, data []byte) error {
	fullPath := filepath.Join(s.basePath, path)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Retrieve retrieves data from the given path
func (s *Service) Retrieve(path string) ([]byte, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// Delete deletes data at the given path
func (s *Service) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists at the given path
func (s *Service) Exists(path string) bool {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// List lists files in the given directory
func (s *Service) List(path string) ([]string, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	return files, nil
}

// Copy copies a file from src to dst
func (s *Service) Copy(src, dst string) error {
	srcPath := filepath.Join(s.basePath, src)
	dstPath := filepath.Join(s.basePath, dst)

	// Create destination directory
	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy data
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// Move moves a file from src to dst
func (s *Service) Move(src, dst string) error {
	srcPath := filepath.Join(s.basePath, src)
	dstPath := filepath.Join(s.basePath, dst)

	// Create destination directory
	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move file
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// Size returns the size of a file
func (s *Service) Size(path string) (int64, error) {
	fullPath := filepath.Join(s.basePath, path)
	
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return info.Size(), nil
}

// CreateDirectory creates a directory
func (s *Service) CreateDirectory(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}
