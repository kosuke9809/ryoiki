package config

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetConfigDir returns the XDG config directory for ryoiki
func GetConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	ryoikiDir := filepath.Join(configDir, "ryoiki")
	return ryoikiDir, nil
}

// GetWorkspacesDir returns the directory where workspace metadata files are stored
func GetWorkspacesDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	workspacesDir := filepath.Join(configDir, "workspaces")
	return workspacesDir, nil
}

// GetMetadataFilePath returns the full path to the metadata file for a given repository root
func GetMetadataFilePath(repoRoot string) (string, error) {
	workspacesDir, err := GetWorkspacesDir()
	if err != nil {
		return "", err
	}

	repoHash := generateRepoHash(repoRoot)
	filename := fmt.Sprintf("%s.json", repoHash)

	return filepath.Join(workspacesDir, filename), nil
}

// generateRepoHash creates a 16-character hash from the repository root path
func generateRepoHash(repoRoot string) string {
	// Get absolute path to ensure consistent hashing
	absPath, err := filepath.Abs(repoRoot)
	if err != nil {
		// Fallback to original path if absolute path fails
		absPath = repoRoot
	}

	hash := sha256.Sum256([]byte(absPath))
	// Return first 16 characters as hexadecimal
	return fmt.Sprintf("%x", hash)[:16]
}

// EnsureConfigDirs creates the config directories if they don't exist
func EnsureConfigDirs() error {
	workspacesDir, err := GetWorkspacesDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(workspacesDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directories: %w", err)
	}

	return nil
}

// GetDefaultWorkspacePath returns the default workspace directory:
// ~/.ryoiki/<repo>-<repohash>/<workspace>
func GetDefaultWorkspacePath(repoRoot, workspaceName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		absRoot = repoRoot
	}

	repoName := filepath.Base(absRoot)
	if repoName == "." || repoName == string(filepath.Separator) || repoName == "" {
		repoName = "repo"
	}
	repoName = sanitizePathComponent(repoName)
	repoScope := fmt.Sprintf("%s-%s", repoName, generateRepoHash(absRoot))
	wsName := sanitizePathComponent(workspaceName)

	return filepath.Join(homeDir, ".ryoiki", repoScope, wsName), nil
}

func sanitizePathComponent(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "workspace"
	}
	s = strings.ReplaceAll(s, string(filepath.Separator), "_")
	return s
}
