package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/shell"
)

var initCmd = &cobra.Command{
	Use:   "init <shell>",
	Short: "Print shell integration script",
	Long:  "Print a shell function for your rc file. Enables 'ryoiki switch' to cd directly.\n\nUsage:\n  eval \"$(ryoiki init zsh)\"",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		script, err := shell.InitScript(args[0])
		if err != nil {
			return err
		}
		fmt.Print(script)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
