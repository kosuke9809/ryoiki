package jj

import (
	"encoding/json"
	"fmt"
	"strings"
)

type WorkspaceService struct {
	executor Executor
}

func NewWorkspaceService(executor Executor) *WorkspaceService {
	return &WorkspaceService{executor: executor}
}

// List returns all workspaces in the repository
func (ws *WorkspaceService) List() ([]Workspace, error) {
	output, err := ws.executor.Execute([]string{"workspace", "list", "--template", "json(self) ++ \"\\n\""})
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var workspaces []Workspace

	for _, line := range lines {
		if line == "" {
			continue
		}

		var workspace Workspace
		if err := json.Unmarshal([]byte(line), &workspace); err != nil {
			return nil, fmt.Errorf("failed to parse workspace JSON: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

// Root returns the workspace root directory path
func (ws *WorkspaceService) Root() (string, error) {
	output, err := ws.executor.Execute([]string{"workspace", "root"})
	if err != nil {
		return "", fmt.Errorf("failed to get workspace root: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Add creates a new workspace with the given name
func (ws *WorkspaceService) Add(name string) error {
	_, err := ws.executor.Execute([]string{"workspace", "add", name})
	if err != nil {
		return fmt.Errorf("failed to add workspace %q: %w", name, err)
	}

	return nil
}

// Forget removes a workspace from the repository
func (ws *WorkspaceService) Forget(name string) error {
	_, err := ws.executor.Execute([]string{"workspace", "forget", name})
	if err != nil {
		return fmt.Errorf("failed to forget workspace %q: %w", name, err)
	}

	return nil
}
