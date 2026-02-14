package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/display"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  "List all jj workspaces with their paths and purposes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

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

		// Build display info
		var infos []display.WorkspaceInfo
		for _, w := range workspaces {
			info := display.WorkspaceInfo{
				Name:     w.Name,
				ChangeID: w.Target.ChangeID,
				CommitID: w.Target.CommitID,
			}

			if meta, ok := metadataMap[w.Name]; ok {
				info.Path = meta.Path
				info.Purpose = meta.Purpose
				info.CreatedAt = meta.CreatedAt
				info.UpdatedAt = meta.UpdatedAt
			}

			// Default workspace path is the repo root
			if w.Name == "default" && info.Path == "" {
				info.Path = root
			}

			// Mark current workspace
			if info.Path == currentRoot {
				info.IsCurrent = true
			}

			infos = append(infos, info)
		}

		if jsonOutput {
			return display.PrintJSON(os.Stdout, infos)
		}

		display.PrintListTable(os.Stdout, infos)
		return nil
	},
}

func init() {
	listCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(listCmd)
}
