package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestRequireInteractiveTerminal(t *testing.T) {
	output := os.Stdout

	tests := []struct {
		name        string
		stdinOK     bool
		outputOK    bool
		wantErr     bool
		wantContain string
	}{
		{
			name:     "interactive stdin and output",
			stdinOK:  true,
			outputOK: true,
			wantErr:  false,
		},
		{
			name:        "non-interactive stdin",
			stdinOK:     false,
			outputOK:    true,
			wantErr:     true,
			wantContain: "command ryoiki tenkai",
		},
		{
			name:        "non-interactive stdout",
			stdinOK:     true,
			outputOK:    false,
			wantErr:     true,
			wantContain: "interactive terminal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Distinguish stdin/output behavior while keeping the function signature.
			isTerm := func(f *os.File) bool {
				if f == nil {
					return false
				}
				if f == output {
					return tt.outputOK
				}
				return tt.stdinOK
			}

			err := requireInteractiveTerminal(os.Stdin, output, isTerm)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantContain != "" && (err == nil || !strings.Contains(err.Error(), tt.wantContain)) {
				t.Fatalf("error %q should contain %q", err, tt.wantContain)
			}
		})
	}
}

func TestIsInteractiveTerminalNilFile(t *testing.T) {
	if isInteractiveTerminal(nil) {
		t.Fatal("expected nil file to be non-interactive")
	}
}

func TestWriteSwitchPath(t *testing.T) {
	t.Run("missing switch file env", func(t *testing.T) {
		t.Setenv(tenkaiSwitchFileEnv, "")
		err := writeSwitchPath("/tmp/ws")
		if err == nil || !strings.Contains(err.Error(), "updated shell integration") {
			t.Fatalf("expected shell integration error, got: %v", err)
		}
	})

	t.Run("writes switch path", func(t *testing.T) {
		f, err := os.CreateTemp("", "ryoiki-switch-path")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		path := f.Name()
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close temp file: %v", err)
		}
		t.Cleanup(func() { _ = os.Remove(path) })

		t.Setenv(tenkaiSwitchFileEnv, path)
		if err := writeSwitchPath("/tmp/ws"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(data) != "/tmp/ws" {
			t.Fatalf("unexpected content: %q", string(data))
		}
	})
}
