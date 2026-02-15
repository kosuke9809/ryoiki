package jj

import (
	"errors"
	"testing"
)

// MockExecutor implements the Executor interface for testing
type MockExecutor struct {
	outputs map[string][]byte
	errors  map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		outputs: make(map[string][]byte),
		errors:  make(map[string]error),
	}
}

func (m *MockExecutor) Execute(args []string) ([]byte, error) {
	key := argsToKey(args)
	if err, exists := m.errors[key]; exists {
		return nil, err
	}
	if output, exists := m.outputs[key]; exists {
		return output, nil
	}
	return nil, errors.New("no mock response configured for: " + key)
}

func (m *MockExecutor) SetOutput(args []string, output []byte) {
	m.outputs[argsToKey(args)] = output
}

func (m *MockExecutor) SetError(args []string, err error) {
	m.errors[argsToKey(args)] = err
}

func argsToKey(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

func TestWorkspaceService_List(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult []Workspace
		expectedError  bool
	}{
		{
			name: "single workspace",
			mockOutput: `{"name":"default","target":{"commit_id":"54c324e86d653a79a7d4bdde0484713cbd1506be","parents":["471696e9eb44026b47b24cf0fcb34733424ba6b6"],"change_id":"xtsuszwwuyvkunrknnnzvkntwvpksvvl","description":"","author":{"name":"user","email":"user@example.com","timestamp":"2026-02-14T13:13:20.108+09:00"},"committer":{"name":"user","email":"user@example.com","timestamp":"2026-02-14T13:13:20+09:00"}}}`,
			expectedResult: []Workspace{
				{
					Name: "default",
					Target: Target{
						CommitID:    "54c324e86d653a79a7d4bdde0484713cbd1506be",
						Parents:     []string{"471696e9eb44026b47b24cf0fcb34733424ba6b6"},
						ChangeID:    "xtsuszwwuyvkunrknnnzvkntwvpksvvl",
						Description: "",
						Author: User{
							Name:  "user",
							Email: "user@example.com",
						},
						Committer: User{
							Name:  "user",
							Email: "user@example.com",
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "multiple workspaces",
			mockOutput: `{"name":"default","target":{"commit_id":"abc123","parents":[],"change_id":"change1","description":"","author":{"name":"user","email":"user@example.com","timestamp":"2026-02-14T13:13:20+09:00"},"committer":{"name":"user","email":"user@example.com","timestamp":"2026-02-14T13:13:20+09:00"}}}
{"name":"feature","target":{"commit_id":"def456","parents":[],"change_id":"change2","description":"feature branch","author":{"name":"user","email":"user@example.com","timestamp":"2026-02-14T13:13:20+09:00"},"committer":{"name":"user","email":"user@example.com","timestamp":"2026-02-14T13:13:20+09:00"}}}`,
			expectedResult: []Workspace{
				{
					Name: "default",
					Target: Target{
						CommitID:    "abc123",
						Parents:     []string{},
						ChangeID:    "change1",
						Description: "",
						Author: User{
							Name:  "user",
							Email: "user@example.com",
						},
						Committer: User{
							Name:  "user",
							Email: "user@example.com",
						},
					},
				},
				{
					Name: "feature",
					Target: Target{
						CommitID:    "def456",
						Parents:     []string{},
						ChangeID:    "change2",
						Description: "feature branch",
						Author: User{
							Name:  "user",
							Email: "user@example.com",
						},
						Committer: User{
							Name:  "user",
							Email: "user@example.com",
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name:           "command execution error",
			mockError:      errors.New("jj command failed"),
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:           "invalid JSON",
			mockOutput:     `invalid json`,
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			if tt.mockError != nil {
				mockExec.SetError([]string{"workspace", "list", "--template", "json(self) ++ \"\\n\""}, tt.mockError)
			} else {
				mockExec.SetOutput([]string{"workspace", "list", "--template", "json(self) ++ \"\\n\""}, []byte(tt.mockOutput))
			}

			service := NewWorkspaceService(mockExec)
			result, err := service.List()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Compare results (ignoring timestamp fields for simplicity)
			if len(result) != len(tt.expectedResult) {
				t.Errorf("expected %d workspaces, got %d", len(tt.expectedResult), len(result))
				return
			}

			for i, expected := range tt.expectedResult {
				if result[i].Name != expected.Name {
					t.Errorf("workspace %d: expected name %q, got %q", i, expected.Name, result[i].Name)
				}
				if result[i].Target.CommitID != expected.Target.CommitID {
					t.Errorf("workspace %d: expected commit ID %q, got %q", i, expected.Target.CommitID, result[i].Target.CommitID)
				}
			}
		})
	}
}

func TestWorkspaceService_Root(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult string
		expectedError  bool
	}{
		{
			name:           "valid root path",
			mockOutput:     "/home/user/project\n",
			expectedResult: "/home/user/project",
			expectedError:  false,
		},
		{
			name:           "root path without newline",
			mockOutput:     "/home/user/project",
			expectedResult: "/home/user/project",
			expectedError:  false,
		},
		{
			name:          "command execution error",
			mockError:     errors.New("jj workspace root failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			if tt.mockError != nil {
				mockExec.SetError([]string{"workspace", "root", "--name", "default"}, tt.mockError)
			} else {
				mockExec.SetOutput([]string{"workspace", "root", "--name", "default"}, []byte(tt.mockOutput))
			}

			service := NewWorkspaceService(mockExec)
			result, err := service.Root()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expectedResult {
				t.Errorf("expected %q, got %q", tt.expectedResult, result)
			}
		})
	}
}

func TestWorkspaceService_Add(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		opts          AddOptions
		expectedArgs  []string
		mockError     error
		expectedError bool
	}{
		{
			name:         "successful add with path only",
			path:         "./feature-branch",
			opts:         AddOptions{},
			expectedArgs: []string{"workspace", "add", "./feature-branch"},
		},
		{
			name:         "add with custom name",
			path:         "./ws-auth",
			opts:         AddOptions{Name: "auth"},
			expectedArgs: []string{"workspace", "add", "./ws-auth", "--name", "auth"},
		},
		{
			name:         "add with revision",
			path:         "./ws-fix",
			opts:         AddOptions{Revision: "main"},
			expectedArgs: []string{"workspace", "add", "./ws-fix", "--revision", "main"},
		},
		{
			name:         "add with all options",
			path:         "./ws-full",
			opts:         AddOptions{Name: "full", Revision: "abc123"},
			expectedArgs: []string{"workspace", "add", "./ws-full", "--name", "full", "--revision", "abc123"},
		},
		{
			name:          "command execution error",
			path:          "./invalid",
			opts:          AddOptions{},
			expectedArgs:  []string{"workspace", "add", "./invalid"},
			mockError:     errors.New("workspace creation failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			if tt.mockError != nil {
				mockExec.SetError(tt.expectedArgs, tt.mockError)
			} else {
				mockExec.SetOutput(tt.expectedArgs, []byte("Created workspace"))
			}

			service := NewWorkspaceService(mockExec)
			err := service.Add(tt.path, tt.opts)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWorkspaceService_Forget(t *testing.T) {
	tests := []struct {
		name          string
		workspaceName string
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful forget",
			workspaceName: "old-feature",
			expectedError: false,
		},
		{
			name:          "command execution error",
			workspaceName: "nonexistent",
			mockError:     errors.New("No such workspace: nonexistent"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			if tt.mockError != nil {
				mockExec.SetError([]string{"workspace", "forget", tt.workspaceName}, tt.mockError)
			} else {
				mockExec.SetOutput([]string{"workspace", "forget", tt.workspaceName}, []byte(""))
			}

			service := NewWorkspaceService(mockExec)
			err := service.Forget(tt.workspaceName)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}