package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/jj"
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Create a new workspace",
	Long:  "Create a new jj workspace at the given path with optional metadata.\nIf path is omitted and --name is given, defaults to .ryoiki/<name>.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		revision, _ := cmd.Flags().GetString("revision")
		purpose, _ := cmd.Flags().GetString("purpose")

		ws, store, root, err := initServices()
		if err != nil {
			return err
		}

		// Determine path
		var path string
		if len(args) > 0 {
			path = args[0]
		} else if name != "" {
			path = filepath.Join(root, ".ryoiki", name)
		} else {
			return fmt.Errorf("either <path> argument or --name flag is required")
		}

		// Add workspace via jj
		if err := ws.Add(path, jj.AddOptions{Name: name, Revision: revision}); err != nil {
			return err
		}

		// Determine workspace name for metadata
		wsName := name
		if wsName == "" {
			wsName = filepath.Base(path)
		}

		// Resolve absolute path for storage
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		// Store metadata
		if err := store.AddWorkspace(wsName, absPath, purpose); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}

		fmt.Printf("Created workspace %q at %s\n", wsName, absPath)
		if purpose != "" {
			fmt.Printf("Purpose: %s\n", purpose)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Workspace name (default: basename of path)")
	addCmd.Flags().StringP("revision", "r", "", "Parent revision")
	addCmd.Flags().StringP("purpose", "p", "", "Purpose description")
	rootCmd.AddCommand(addCmd)
}
