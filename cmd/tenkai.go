package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/tui"
)

const tenkaiSwitchFileEnv = "RYOIKI_TENKAI_SWITCH_FILE"

var tenkaiCmd = &cobra.Command{
	Use:   "tenkai",
	Short: "Launch interactive TUI",
	Long:  "Launch the interactive TUI for workspace management.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireInteractiveTerminal(os.Stdin, os.Stdout, isInteractiveTerminal); err != nil {
			return err
		}

		ws, store, root, err := initServices()
		if err != nil {
			return err
		}
		app := tui.NewApp(ws, store, root)
		p := tea.NewProgram(app, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		if finalApp, ok := finalModel.(tui.App); ok && finalApp.SwitchPath != "" {
			return writeSwitchPath(finalApp.SwitchPath)
		}
		return nil
	},
}

func writeSwitchPath(path string) error {
	switchFile := os.Getenv(tenkaiSwitchFileEnv)
	if switchFile == "" {
		return fmt.Errorf(
			"tenkai switch requires updated shell integration. Re-run `ryoiki init <shell>` and reload your shell (e.g. `source ~/.zshrc`)",
		)
	}
	if err := os.WriteFile(switchFile, []byte(path), 0600); err != nil {
		return fmt.Errorf("failed to write switch path: %w", err)
	}
	return nil
}

func requireInteractiveTerminal(stdin, output *os.File, isTerminal func(*os.File) bool) error {
	if !isTerminal(stdin) || !isTerminal(output) {
		return fmt.Errorf(
			"tenkai requires an interactive terminal (stdin/stdout must be a TTY). " +
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
