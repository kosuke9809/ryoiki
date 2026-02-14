package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/display"
)

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show workspace details",
	Long:  "Show detailed information about a specific workspace.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		ws, store, root, err := initServices()
		if err != nil {
			return err
		}

		// Get workspaces from jj
		workspaces, err := ws.List()
		if err != nil {
			return err
		}

		// Get metadata
		metadataMap, err := store.ListWorkspaces()
		if err != nil {
			return err
		}

		// Get current workspace root for marking current
		currentRoot, _ := ws.Root()

		infos := display.BuildWorkspaceInfos(workspaces, metadataMap, root, currentRoot)

		// Find the target workspace
		var found bool
		for _, info := range infos {
			if info.Name == name {
				display.PrintShowDetail(os.Stdout, info)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("workspace %q not found", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
