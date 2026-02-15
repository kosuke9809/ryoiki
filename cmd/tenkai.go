package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/tui"
)

const tenkaiSwitchCaptureEnv = "RYOIKI_TENKAI_SWITCH_CAPTURE"

var tenkaiCmd = &cobra.Command{
	Use:   "tenkai",
	Short: "Launch interactive TUI",
	Long:  "Launch the interactive TUI for workspace management.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireInteractiveTerminal(os.Stdin, os.Stderr, isInteractiveTerminal); err != nil {
			return err
		}

		ws, store, root, err := initServices()
		if err != nil {
			return err
		}
		app := tui.NewApp(ws, store, root)
		// Render TUI to stderr so stdout can be reserved for switch path output.
		p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		if finalApp, ok := finalModel.(tui.App); ok && finalApp.SwitchPath != "" {
			if !switchCaptureEnabled() {
				return fmt.Errorf(
					"tenkai switch requires updated shell integration. Re-run `ryoiki init <shell>` and reload your shell (e.g. `source ~/.zshrc`)",
				)
			}
			fmt.Print(finalApp.SwitchPath)
		}
		return nil
	},
}

func switchCaptureEnabled() bool {
	return os.Getenv(tenkaiSwitchCaptureEnv) == "1"
}

func requireInteractiveTerminal(stdin, output *os.File, isTerminal func(*os.File) bool) error {
	if !isTerminal(stdin) || !isTerminal(output) {
		return fmt.Errorf(
			"tenkai requires an interactive terminal (stdin/stderr must be a TTY). " +
				"If your shell wrapper is intercepting this command, run `command ryoiki tenkai` or re-run `ryoiki init <shell>`",
		)
	}
	return nil
}

func isInteractiveTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func init() {
	rootCmd.AddCommand(tenkaiCmd)
}
