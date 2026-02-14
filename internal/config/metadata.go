package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// WorkspaceMetadata holds metadata for a specific workspace
type WorkspaceMetadata struct {
	Path      string    `json:"path,omitempty"`
	Purpose   string    `json:"purpose"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RepositoryMetadata holds metadata for an entire repository
type RepositoryMetadata struct {
	RepoPath   string                       `json:"repo_path"`
	Workspaces map[string]WorkspaceMetadata `json:"workspaces"`
}

// MetadataStore manages workspace metadata persistence
type MetadataStore struct {
	repoRoot string
}

// NewMetadataStore creates a new MetadataStore for the given repository root
func NewMetadataStore(repoRoot string) *MetadataStore {
	return &MetadataStore{repoRoot: repoRoot}
}

// Load reads the repository metadata from disk
func (ms *MetadataStore) Load() (*RepositoryMetadata, error) {
	filePath, err := GetMetadataFilePath(ms.repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return empty metadata if file doesn't exist
		return &RepositoryMetadata{
			RepoPath:   ms.repoRoot,
			Workspaces: make(map[string]WorkspaceMetadata),
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata RepositoryMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	// Ensure workspaces map is initialized
	if metadata.Workspaces == nil {
		metadata.Workspaces = make(map[string]WorkspaceMetadata)
	}

	return &metadata, nil
}

// Save writes the repository metadata to disk
func (ms *MetadataStore) Save(metadata *RepositoryMetadata) error {
	if err := EnsureConfigDirs(); err != nil {
		return fmt.Errorf("failed to ensure config directories: %w", err)
	}

	filePath, err := GetMetadataFilePath(ms.repoRoot)
	if err != nil {
		return fmt.Errorf("failed to get metadata file path: %w", err)
	}

	// Ensure repo path is set correctly
	metadata.RepoPath = ms.repoRoot

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// SetPurpose sets the purpose for a specific workspace
func (ms *MetadataStore) SetPurpose(workspaceName, purpose string) error {
	metadata, err := ms.Load()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	now := time.Now()
	if existing, exists := metadata.Workspaces[workspaceName]; exists {
		// Update existing workspace metadata
		existing.Purpose = purpose
		existing.UpdatedAt = now
		metadata.Workspaces[workspaceName] = existing
	} else {
		// Create new workspace metadata
		metadata.Workspaces[workspaceName] = WorkspaceMetadata{
			Purpose:   purpose,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return ms.Save(metadata)
}

// Remove removes metadata for a specific workspace
func (ms *MetadataStore) Remove(workspaceName string) error {
	metadata, err := ms.Load()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	delete(metadata.Workspaces, workspaceName)
	return ms.Save(metadata)
}

// GetPurpose retrieves the purpose for a specific workspace
func (ms *MetadataStore) GetPurpose(workspaceName string) (string, error) {
	metadata, err := ms.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load metadata: %w", err)
	}

	if workspace, exists := metadata.Workspaces[workspaceName]; exists {
		return workspace.Purpose, nil
	}

	return "", nil // No purpose set
}

// ListWorkspaces returns all workspace metadata
func (ms *MetadataStore) ListWorkspaces() (map[string]WorkspaceMetadata, error) {
	metadata, err := ms.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	return metadata.Workspaces, nil
}

// AddWorkspace creates a new workspace metadata entry with path and purpose
func (ms *MetadataStore) AddWorkspace(name, path, purpose string) error {
	metadata, err := ms.Load()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	now := time.Now()
	metadata.Workspaces[name] = WorkspaceMetadata{
		Path:      path,
		Purpose:   purpose,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return ms.Save(metadata)
}

// GetPath retrieves the path for a specific workspace
func (ms *MetadataStore) GetPath(workspaceName string) (string, error) {
	metadata, err := ms.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load metadata: %w", err)
	}

	if workspace, exists := metadata.Workspaces[workspaceName]; exists {
		return workspace.Path, nil
	}

	return "", nil
}