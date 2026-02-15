package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Output workspace path for switching",
	Long:  "Output the path of the specified workspace. Use with cd: cd $(ryoiki switch <name>)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		_, store, root, err := initServices()
		if err != nil {
			return err
		}

		// For default workspace, return repo root
		if name == "default" {
			fmt.Print(root)
			return nil
		}

		// Resolve path with fallback chain
		path, err := store.ResolveWorkspacePath(name)
		if err != nil {
			return err
		}

		fmt.Print(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
