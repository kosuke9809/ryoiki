package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestRequireInteractiveTerminal(t *testing.T) {
	file := os.Stdout

	tests := []struct {
		name        string
		stdinOK     bool
		stdoutOK    bool
		wantErr     bool
		wantContain string
	}{
		{
			name:     "interactive stdin and stdout",
			stdinOK:  true,
			stdoutOK: true,
			wantErr:  false,
		},
		{
			name:        "non-interactive stdin",
			stdinOK:     false,
			stdoutOK:    true,
			wantErr:     true,
			wantContain: "command ryoiki tenkai",
		},
		{
			name:        "non-interactive stdout",
			stdinOK:     true,
			stdoutOK:    false,
			wantErr:     true,
			wantContain: "interactive terminal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Distinguish stdin/stdout behavior while keeping the function signature.
			isTerm := func(f *os.File) bool {
				if f == nil {
					return false
				}
				if f == file {
					return tt.stdoutOK
				}
				return tt.stdinOK
			}

			err := requireInteractiveTerminal(os.Stdin, file, isTerm)
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
