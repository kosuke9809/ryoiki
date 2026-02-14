package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tenkaiCmd = &cobra.Command{
	Use:   "tenkai",
	Short: "Launch interactive TUI",
	Long:  "Launch the interactive TUI for workspace management.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TUI mode is not yet implemented. Coming in Phase 2.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tenkaiCmd)
}
