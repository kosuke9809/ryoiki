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

// AddOptions holds optional parameters for workspace creation
type AddOptions struct {
	Name     string
	Revision string
}

// Add creates a new workspace at the given path
func (ws *WorkspaceService) Add(path string, opts AddOptions) error {
	args := []string{"workspace", "add", path}
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	if opts.Revision != "" {
		args = append(args, "--revision", opts.Revision)
	}
	_, err := ws.executor.Execute(args)
	if err != nil {
		return fmt.Errorf("failed to add workspace at %q: %w", path, err)
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
