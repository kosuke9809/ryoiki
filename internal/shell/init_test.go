package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitScript(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr bool
		contain string
	}{
		{"zsh", false, `\command ryoiki "$@"`},
		{"bash", false, `command ryoiki "$@"`},
		{"fish", false, "command ryoiki $argv"},
		{"powershell", true, ""},
		{"", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			script, err := InitScript(tt.shell)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(script, tt.contain) {
				t.Errorf("script does not contain %q:\n%s", tt.contain, script)
			}
		})
	}
}

func TestInitScriptDefinesFunction(t *testing.T) {
	shells := []struct {
		name   string
		prefix string
	}{
		{"zsh", "ryoiki()"},
		{"bash", "ryoiki()"},
		{"fish", "function ryoiki"},
	}

	for _, s := range shells {
		t.Run(s.name, func(t *testing.T) {
			script, err := InitScript(s.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(script, s.prefix) {
				t.Errorf("script does not define function with %q:\n%s", s.prefix, script)
			}
		})
	}
}

func TestConfigFilePath(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr bool
		suffix  string
	}{
		{"zsh", false, ".zshrc"},
		{"bash", false, ".bashrc"},
		{"fish", false, filepath.Join(".config", "fish", "config.fish")},
		{"powershell", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			path, err := ConfigFilePath(tt.shell)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasSuffix(path, tt.suffix) {
				t.Errorf("path %q should end with %q", path, tt.suffix)
			}
		})
	}
}

func TestEvalLine(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr bool
		contain string
	}{
		{"zsh", false, `eval "$(ryoiki init zsh --print)"`},
		{"bash", false, `eval "$(ryoiki init bash --print)"`},
		{"fish", false, "ryoiki init fish --print | source"},
		{"powershell", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			line, err := EvalLine(tt.shell)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if line != tt.contain {
				t.Errorf("got %q, want %q", line, tt.contain)
			}
		})
	}
}

func TestInitScriptInterceptsTenkai(t *testing.T) {
	tests := []struct {
		shell   string
		contain []string
	}{
		{"zsh", []string{`"switch"`, `"tenkai"`, "RYOIKI_TENKAI_SWITCH_CAPTURE=1", "RYOIKI_SHELL_HOOK_VERSION=2"}},
		{"bash", []string{`"switch"`, `"tenkai"`, "RYOIKI_TENKAI_SWITCH_CAPTURE=1", "RYOIKI_SHELL_HOOK_VERSION=2"}},
		{"fish", []string{`"switch"`, `"tenkai"`, "RYOIKI_TENKAI_SWITCH_CAPTURE=1", "RYOIKI_SHELL_HOOK_VERSION 2"}},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			script, err := InitScript(tt.shell)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, contain := range tt.contain {
				if !strings.Contains(script, contain) {
					t.Errorf("script missing %q:\n%s", contain, script)
				}
			}
		})
	}
}

func TestWriteToConfig(t *testing.T) {
	// Create a temp home directory
	tmpHome, err := os.MkdirTemp("", "ryoiki-shell-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpHome); err != nil {
			t.Errorf("failed to cleanup temp home: %v", err)
		}
	})

	// Override HOME for testing
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Errorf("failed to restore HOME: %v", err)
		}
	})

	t.Run("writes to new file", func(t *testing.T) {
		already, err := WriteToConfig("zsh")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if already {
			t.Error("expected already=false for new file")
		}

		content, err := os.ReadFile(filepath.Join(tmpHome, ".zshrc"))
		if err != nil {
			t.Fatalf("failed to read .zshrc: %v", err)
		}
		if !strings.Contains(string(content), `eval "$(ryoiki init zsh --print)"`) {
			t.Errorf("config file missing eval line: %s", content)
		}
	})

	t.Run("detects already configured", func(t *testing.T) {
		already, err := WriteToConfig("zsh")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !already {
			t.Error("expected already=true for second call")
		}
	})

	t.Run("creates fish config directory", func(t *testing.T) {
		already, err := WriteToConfig("fish")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if already {
			t.Error("expected already=false for new fish config")
		}

		fishConfig := filepath.Join(tmpHome, ".config", "fish", "config.fish")
		content, err := os.ReadFile(fishConfig)
		if err != nil {
			t.Fatalf("failed to read config.fish: %v", err)
		}
		if !strings.Contains(string(content), "ryoiki init fish --print | source") {
			t.Errorf("fish config missing eval line: %s", content)
		}
	})
}
