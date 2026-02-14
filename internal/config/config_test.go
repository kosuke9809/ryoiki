package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name           string
		xdgConfigHome  string
		expectedSuffix string
	}{
		{
			name:           "uses XDG_CONFIG_HOME when set",
			xdgConfigHome:  "/custom/config",
			expectedSuffix: "custom/config/ryoiki",
		},
		{
			name:           "uses default ~/.config when XDG_CONFIG_HOME not set",
			xdgConfigHome:  "",
			expectedSuffix: ".config/ryoiki",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

			// Set test environment
			if tt.xdgConfigHome == "" {
				os.Unsetenv("XDG_CONFIG_HOME")
			} else {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			}

			result, err := GetConfigDir()
			if err != nil {
				t.Errorf("GetConfigDir() error = %v", err)
				return
			}

			if !filepath.IsAbs(result) {
				t.Errorf("GetConfigDir() should return absolute path, got %q", result)
			}

			if !strings.HasSuffix(result, tt.expectedSuffix) {
				t.Errorf("GetConfigDir() should end with %q, got %q", tt.expectedSuffix, result)
			}
		})
	}
}

func TestGenerateRepoHash(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		expected int // length
	}{
		{
			name:     "generates 16 character hash",
			repoPath: "/home/user/project",
			expected: 16,
		},
		{
			name:     "consistent hash for same path",
			repoPath: "/test/path",
			expected: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result1 := generateRepoHash(tt.repoPath)
			result2 := generateRepoHash(tt.repoPath)

			if len(result1) != tt.expected {
				t.Errorf("generateRepoHash() length = %d, want %d", len(result1), tt.expected)
			}

			if result1 != result2 {
				t.Errorf("generateRepoHash() should be consistent, got %q and %q", result1, result2)
			}

			// Different paths should produce different hashes
			different := generateRepoHash(tt.repoPath + "/different")
			if result1 == different {
				t.Errorf("generateRepoHash() should produce different hashes for different paths")
			}
		})
	}
}

func TestGetMetadataFilePath(t *testing.T) {
	repoPath := "/test/repo"
	result, err := GetMetadataFilePath(repoPath)
	if err != nil {
		t.Errorf("GetMetadataFilePath() error = %v", err)
		return
	}

	if !strings.HasSuffix(result, ".json") {
		t.Errorf("GetMetadataFilePath() should end with .json, got %q", result)
	}

	if !filepath.IsAbs(result) {
		t.Errorf("GetMetadataFilePath() should return absolute path, got %q", result)
	}
}

func TestMetadataStore_LoadSave(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ryoiki-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override XDG_CONFIG_HOME for testing
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	repoPath := "/test/repo"
	store := NewMetadataStore(repoPath)

	// Test loading non-existent metadata (should return empty)
	metadata, err := store.Load()
	if err != nil {
		t.Errorf("Load() error = %v", err)
		return
	}

	if metadata.RepoPath != repoPath {
		t.Errorf("Load() repo path = %q, want %q", metadata.RepoPath, repoPath)
	}

	if len(metadata.Workspaces) != 0 {
		t.Errorf("Load() should return empty workspaces, got %d", len(metadata.Workspaces))
	}

	// Add some test data
	now := time.Now()
	metadata.Workspaces["test-workspace"] = WorkspaceMetadata{
		Purpose:   "testing workspace",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test saving
	if err := store.Save(metadata); err != nil {
		t.Errorf("Save() error = %v", err)
		return
	}

	// Test loading saved data
	loadedMetadata, err := store.Load()
	if err != nil {
		t.Errorf("Load() after save error = %v", err)
		return
	}

	if loadedMetadata.RepoPath != repoPath {
		t.Errorf("Loaded repo path = %q, want %q", loadedMetadata.RepoPath, repoPath)
	}

	if len(loadedMetadata.Workspaces) != 1 {
		t.Errorf("Loaded workspaces count = %d, want 1", len(loadedMetadata.Workspaces))
	}

	workspace, exists := loadedMetadata.Workspaces["test-workspace"]
	if !exists {
		t.Errorf("Expected workspace 'test-workspace' not found")
		return
	}

	if workspace.Purpose != "testing workspace" {
		t.Errorf("Workspace purpose = %q, want %q", workspace.Purpose, "testing workspace")
	}
}

func TestMetadataStore_SetPurpose(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ryoiki-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override XDG_CONFIG_HOME for testing
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	repoPath := "/test/repo"
	store := NewMetadataStore(repoPath)

	// Test setting purpose for new workspace
	workspaceName := "feature-branch"
	purpose := "implementing new feature"
	
	if err := store.SetPurpose(workspaceName, purpose); err != nil {
		t.Errorf("SetPurpose() error = %v", err)
		return
	}

	// Verify purpose was set
	retrievedPurpose, err := store.GetPurpose(workspaceName)
	if err != nil {
		t.Errorf("GetPurpose() error = %v", err)
		return
	}

	if retrievedPurpose != purpose {
		t.Errorf("GetPurpose() = %q, want %q", retrievedPurpose, purpose)
	}

	// Test updating existing workspace purpose
	newPurpose := "updated feature description"
	if err := store.SetPurpose(workspaceName, newPurpose); err != nil {
		t.Errorf("SetPurpose() update error = %v", err)
		return
	}

	updatedPurpose, err := store.GetPurpose(workspaceName)
	if err != nil {
		t.Errorf("GetPurpose() after update error = %v", err)
		return
	}

	if updatedPurpose != newPurpose {
		t.Errorf("GetPurpose() after update = %q, want %q", updatedPurpose, newPurpose)
	}
}

func TestMetadataStore_Remove(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ryoiki-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override XDG_CONFIG_HOME for testing
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	repoPath := "/test/repo"
	store := NewMetadataStore(repoPath)

	// Set up test data
	workspaceName := "temp-workspace"
	if err := store.SetPurpose(workspaceName, "temporary workspace"); err != nil {
		t.Errorf("SetPurpose() setup error = %v", err)
		return
	}

	// Verify workspace exists
	purpose, err := store.GetPurpose(workspaceName)
	if err != nil {
		t.Errorf("GetPurpose() setup error = %v", err)
		return
	}
	if purpose == "" {
		t.Errorf("Expected workspace to exist before removal")
	}

	// Remove workspace
	if err := store.Remove(workspaceName); err != nil {
		t.Errorf("Remove() error = %v", err)
		return
	}

	// Verify workspace is removed
	purpose, err = store.GetPurpose(workspaceName)
	if err != nil {
		t.Errorf("GetPurpose() after remove error = %v", err)
		return
	}
	if purpose != "" {
		t.Errorf("Expected workspace to be removed, but purpose is %q", purpose)
	}
}