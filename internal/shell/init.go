package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const zshScript = `ryoiki() {
    export RYOIKI_SHELL_HOOK_VERSION=2
    if [[ "$1" == "switch" ]]; then
        local dir
        dir="$(\command ryoiki "$@")" && [[ -n "$dir" ]] && builtin cd -- "$dir"
    elif [[ "$1" == "tenkai" ]]; then
        local switch_file cmd_status dir
        switch_file="$(mktemp)"
        cmd_status=0
        RYOIKI_TENKAI_SWITCH_FILE="$switch_file" \command ryoiki "$@" || cmd_status=$?
        if [[ $cmd_status -eq 0 ]]; then
            dir="$(cat "$switch_file")"
            [[ -n "$dir" ]] && builtin cd -- "$dir"
        fi
        rm -f -- "$switch_file"
        return $cmd_status
    else
        \command ryoiki "$@"
    fi
}
`

const bashScript = `ryoiki() {
    export RYOIKI_SHELL_HOOK_VERSION=2
    if [[ "$1" == "switch" ]]; then
        local dir
        dir="$(command ryoiki "$@")" && [[ -n "$dir" ]] && builtin cd -- "$dir"
    elif [[ "$1" == "tenkai" ]]; then
        local switch_file status dir
        switch_file="$(mktemp)"
        status=0
        RYOIKI_TENKAI_SWITCH_FILE="$switch_file" command ryoiki "$@" || status=$?
        if [[ $status -eq 0 ]]; then
            dir="$(cat "$switch_file")"
            [[ -n "$dir" ]] && builtin cd -- "$dir"
        fi
        rm -f -- "$switch_file"
        return $status
    else
        command ryoiki "$@"
    fi
}
`

const fishScript = `function ryoiki
    set -gx RYOIKI_SHELL_HOOK_VERSION 2
    if test "$argv[1]" = "switch"
        set -l dir (command ryoiki $argv)
        and test -n "$dir"
        and builtin cd -- $dir
    else if test "$argv[1]" = "tenkai"
        set -l switch_file (mktemp)
        env RYOIKI_TENKAI_SWITCH_FILE=$switch_file command ryoiki $argv
        set -l cmd_status $status
        if test $cmd_status -eq 0
            set -l dir (cat $switch_file)
            and test -n "$dir"
            and builtin cd -- $dir
        end
        rm -f -- $switch_file
        return $cmd_status
    else
        command ryoiki $argv
    end
end
`

// InitScript returns the shell integration script for the given shell.
func InitScript(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshScript, nil
	case "bash":
		return bashScript, nil
	case "fish":
		return fishScript, nil
	default:
		return "", fmt.Errorf("unsupported shell: %q (supported: zsh, bash, fish)", shell)
	}
}

// ConfigFilePath returns the rc file path for the given shell.
func ConfigFilePath(shellName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch shellName {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %q (supported: zsh, bash, fish)", shellName)
	}
}

// EvalLine returns the line to add to the rc file for the given shell.
func EvalLine(shellName string) (string, error) {
	switch shellName {
	case "zsh":
		return `eval "$(ryoiki init zsh --print)"`, nil
	case "bash":
		return `eval "$(ryoiki init bash --print)"`, nil
	case "fish":
		return `ryoiki init fish --print | source`, nil
	default:
		return "", fmt.Errorf("unsupported shell: %q (supported: zsh, bash, fish)", shellName)
	}
}

// WriteToConfig appends the eval line to the shell config file.
// Returns (alreadyConfigured, error).
func WriteToConfig(shellName string) (bool, error) {
	configPath, err := ConfigFilePath(shellName)
	if err != nil {
		return false, err
	}

	evalLine, err := EvalLine(shellName)
	if err != nil {
		return false, err
	}

	// Read existing content (file may not exist yet)
	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	// Check if already configured
	if strings.Contains(string(content), evalLine) {
		return true, nil
	}

	// Ensure parent directory exists (for fish)
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Append eval line
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, fmt.Errorf("failed to open %s: %w", configPath, err)
	}

	// Add newline before if file has content and doesn't end with newline
	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			if closeErr := f.Close(); closeErr != nil {
				return false, fmt.Errorf("failed to write to %s: %v (close error: %v)", configPath, err, closeErr)
			}
			return false, fmt.Errorf("failed to write to %s: %w", configPath, err)
		}
	}

	if _, err := f.WriteString(evalLine + "\n"); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return false, fmt.Errorf("failed to write to %s: %v (close error: %v)", configPath, err, closeErr)
		}
		return false, fmt.Errorf("failed to write to %s: %w", configPath, err)
	}

	if err := f.Close(); err != nil {
		return false, fmt.Errorf("failed to close %s: %w", configPath, err)
	}

	return false, nil
}
