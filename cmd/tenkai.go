package cmd

import (
	"fmt"

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
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		if finalApp, ok := finalModel.(tui.App); ok && finalApp.SwitchPath != "" {
			fmt.Print(finalApp.SwitchPath)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tenkaiCmd)
}
