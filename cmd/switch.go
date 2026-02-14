package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Output workspace path for switching",
	Long:  "Output the path of the specified workspace. Use with cd: cd $(ry switch <name>)",
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

		// Look up path from metadata
		path, err := store.GetPath(name)
		if err != nil {
			return err
		}

		if path == "" {
			return fmt.Errorf("workspace %q not found or path not recorded; register it with 'ry describe %s'", name, name)
		}

		fmt.Print(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
