package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/tui"
)

var tenkaiCmd = &cobra.Command{
	Use:   "tenkai",
	Short: "Launch interactive TUI",
	Long:  "Launch the interactive TUI for workspace management.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, store, root, err := initServices()
		if err != nil {
			return err
		}
		app := tui.NewApp(ws, store, root)
		p := tea.NewProgram(app, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(tenkaiCmd)
}
