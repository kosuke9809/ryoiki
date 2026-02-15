package shell

import "fmt"

const zshScript = `ryoiki() {
    if [[ "$1" == "switch" ]]; then
        local dir
        dir="$(\command ryoiki "$@")" && [[ -n "$dir" ]] && builtin cd -- "$dir"
    else
        \command ryoiki "$@"
    fi
}
`

const bashScript = `ryoiki() {
    if [[ "$1" == "switch" ]]; then
        local dir
        dir="$(command ryoiki "$@")" && [[ -n "$dir" ]] && builtin cd -- "$dir"
    else
        command ryoiki "$@"
    fi
}
`

const fishScript = `function ryoiki
    if test "$argv[1]" = "switch"
        set -l dir (command ryoiki $argv)
        and test -n "$dir"
        and builtin cd -- $dir
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
