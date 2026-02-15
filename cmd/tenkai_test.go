package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestRequireInteractiveTerminal(t *testing.T) {
	output := os.Stderr

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
			name:        "non-interactive stderr",
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
