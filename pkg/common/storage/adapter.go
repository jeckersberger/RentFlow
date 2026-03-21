package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jeckersberger/rentflow/pkg/common/config"
)

// FileMetadata contains metadata about a stored file
type FileMetadata struct {
	TenantID    string
	EntityType  string // "receipt", "equipment", "document", etc.
	EntityID    string
	Filename    string
	ContentType string
	Size        int64
	Checksum    string // SHA-256
	StoredAt    int64
}

// StorageAdapter defines the interface for file storage
// Implementations can use local filesystem, NAS, S3, etc.
type StorageAdapter interface {
	// Store stores a file and returns a file reference
	Store(ctx context.Context, file io.Reader, meta FileMetadata) (string, error)

	// Retrieve retrieves a file by reference
	Retrieve(ctx context.Context, fileRef string) (io.ReadCloser, *FileMetadata, error)

	// Delete deletes a file by reference
	Delete(ctx context.Context, fileRef string) error

	// List lists files for a specific entity
	List(ctx context.Context, entityType, entityID string) ([]FileMetadata, error)

	// Exists checks if a file exists
	Exists(ctx context.Context, fileRef string) (bool, error)

	// Close closes the storage adapter
	Close() error
}

// LocalStorageAdapter stores files on the local filesystem
type LocalStorageAdapter struct {
	basePath string
}

// NewLocalStorageAdapter creates a new local storage adapter
func NewLocalStorageAdapter(basePath string) (*LocalStorageAdapter, error) {
	if basePath == "" {
		return nil, fmt.Errorf("base path cannot be empty")
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorageAdapter{
		basePath: basePath,
	}, nil
}

// Store stores a file on the local filesystem
func (l *LocalStorageAdapter) Store(ctx context.Context, file io.Reader, meta FileMetadata) (string, error) {
	// Create directory structure: basePath/tenantID/entityType/entityID
	dirPath := filepath.Join(l.basePath, meta.TenantID, meta.EntityType, meta.EntityID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file reference
	fileRef := generateFileRef(meta)
	filePath := filepath.Join(dirPath, fileRef)

	// Create file and calculate checksum
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Calculate SHA-256 hash while writing
	hasher := sha256.New()
	mw := io.MultiWriter(f, hasher)

	size, err := io.Copy(mw, file)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	meta.Size = size
	meta.Checksum = fmt.Sprintf("%x", hasher.Sum(nil))

	return fileRef, nil
}

// Retrieve retrieves a file from the local filesystem
func (l *LocalStorageAdapter) Retrieve(ctx context.Context, fileRef string) (io.ReadCloser, *FileMetadata, error) {
	// Search for the file
	var foundPath string
	var foundMeta *FileMetadata

	err := filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == fileRef {
			foundPath = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to search for file: %w", err)
	}

	if foundPath == "" {
		return nil, nil, fmt.Errorf("file not found")
	}

	// Open file
	f, err := os.Open(foundPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Extract metadata from path
	// basePath/tenantID/entityType/entityID/fileRef
	rel, err := filepath.Rel(l.basePath, foundPath)
	if err != nil {
		return f, nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	parts := filepath.SplitList(rel)

	if len(parts) >= 4 {
		foundMeta = &FileMetadata{
			TenantID:   parts[0],
			EntityType: parts[1],
			EntityID:   parts[2],
			Filename:   fileRef,
		}
	}

	return f, foundMeta, nil
}

// Delete deletes a file from the local filesystem
func (l *LocalStorageAdapter) Delete(ctx context.Context, fileRef string) error {
	// Search for and delete the file
	var targetPath string

	err := filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == fileRef {
			targetPath = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to search for file: %w", err)
	}

	if targetPath == "" {
		return fmt.Errorf("file not found")
	}

	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// List lists files for a specific entity
func (l *LocalStorageAdapter) List(ctx context.Context, entityType, entityID string) ([]FileMetadata, error) {
	var results []FileMetadata

	// Search through all tenants for matching entity
	err := filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Extract path components
		rel, err := filepath.Rel(l.basePath, path)
		if err != nil {
			return err
		}
		parts := filepath.SplitList(rel)

		if len(parts) >= 3 && parts[1] == entityType && parts[2] == entityID {
			stat, err := os.Stat(path)
			if err != nil {
				return err
			}
			results = append(results, FileMetadata{
				TenantID:   parts[0],
				EntityType: parts[1],
				EntityID:   parts[2],
				Filename:   info.Name(),
				Size:       stat.Size(),
			})
		}

		return nil
	})

	return results, err
}

// Exists checks if a file exists
func (l *LocalStorageAdapter) Exists(ctx context.Context, fileRef string) (bool, error) {
	var exists bool

	err := filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == fileRef {
			exists = true
			return filepath.SkipDir
		}

		return nil
	})

	return exists, err
}

// Close closes the storage adapter
func (l *LocalStorageAdapter) Close() error {
	return nil
}

// NASStorageAdapter stores files on Network Attached Storage
type NASStorageAdapter struct {
	mountPath string
	nasType   string // "smb", "nfs"
}

// NewNASStorageAdapter creates a new NAS storage adapter
func NewNASStorageAdapter(mountPath, nasType string) (*NASStorageAdapter, error) {
	if mountPath == "" {
		return nil, fmt.Errorf("mount path cannot be empty")
	}

	// Check if mount path exists
	if _, err := os.Stat(mountPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mount path does not exist: %s", mountPath)
	}

	return &NASStorageAdapter{
		mountPath: mountPath,
		nasType:   nasType,
	}, nil
}

// Store stores a file on NAS
func (n *NASStorageAdapter) Store(ctx context.Context, file io.Reader, meta FileMetadata) (string, error) {
	// Similar to LocalStorageAdapter but uses mount path
	dirPath := filepath.Join(n.mountPath, meta.TenantID, meta.EntityType, meta.EntityID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	fileRef := generateFileRef(meta)
	filePath := filepath.Join(dirPath, fileRef)

	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	mw := io.MultiWriter(f, hasher)

	size, err := io.Copy(mw, file)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	meta.Size = size
	meta.Checksum = fmt.Sprintf("%x", hasher.Sum(nil))

	return fileRef, nil
}

// Retrieve retrieves a file from NAS
func (n *NASStorageAdapter) Retrieve(ctx context.Context, fileRef string) (io.ReadCloser, *FileMetadata, error) {
	// Similar to LocalStorageAdapter
	var foundPath string

	err := filepath.Walk(n.mountPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == fileRef {
			foundPath = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to search for file: %w", err)
	}

	if foundPath == "" {
		return nil, nil, fmt.Errorf("file not found")
	}

	f, err := os.Open(foundPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return f, nil, nil
}

// Delete deletes a file from NAS
func (n *NASStorageAdapter) Delete(ctx context.Context, fileRef string) error {
	var targetPath string

	err := filepath.Walk(n.mountPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == fileRef {
			targetPath = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to search for file: %w", err)
	}

	if targetPath == "" {
		return fmt.Errorf("file not found")
	}

	return os.Remove(targetPath)
}

// List lists files on NAS
func (n *NASStorageAdapter) List(ctx context.Context, entityType, entityID string) ([]FileMetadata, error) {
	var results []FileMetadata

	err := filepath.Walk(n.mountPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(n.mountPath, path)
		if err != nil {
			return err
		}
		parts := filepath.SplitList(rel)

		if len(parts) >= 3 && parts[1] == entityType && parts[2] == entityID {
			stat, err := os.Stat(path)
			if err != nil {
				return err
			}
			results = append(results, FileMetadata{
				TenantID:   parts[0],
				EntityType: parts[1],
				EntityID:   parts[2],
				Filename:   info.Name(),
				Size:       stat.Size(),
			})
		}

		return nil
	})

	return results, err
}

// Exists checks if a file exists on NAS
func (n *NASStorageAdapter) Exists(ctx context.Context, fileRef string) (bool, error) {
	var exists bool

	err := filepath.Walk(n.mountPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == fileRef {
			exists = true
			return filepath.SkipDir
		}

		return nil
	})

	return exists, err
}

// Close closes the NAS storage adapter
func (n *NASStorageAdapter) Close() error {
	return nil
}

// NewStorageAdapter creates a storage adapter based on configuration
func NewStorageAdapter(cfg *config.Config) (StorageAdapter, error) {
	if cfg.NASEnabled {
		return NewNASStorageAdapter(cfg.NASMountPath, cfg.NASType)
	}

	return NewLocalStorageAdapter("./storage")
}

// generateFileRef generates a unique file reference
func generateFileRef(meta FileMetadata) string {
	// Use original filename for now
	// In production, generate a UUID to avoid filename conflicts
	return meta.Filename
}
