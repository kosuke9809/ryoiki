package shell

import (
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
