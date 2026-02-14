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

		// Find the target workspace
		var found bool
		var info display.WorkspaceInfo
		for _, w := range workspaces {
			if w.Name == name {
				info = display.WorkspaceInfo{
					Name:        w.Name,
					ChangeID:    w.Target.ChangeID,
					CommitID:    w.Target.CommitID,
					Description: w.Target.Description,
					AuthorName:  w.Target.Author.Name,
					AuthorEmail: w.Target.Author.Email,
				}
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("workspace %q not found", name)
		}

		// Get metadata
		metadataMap, err := store.ListWorkspaces()
		if err != nil {
			return err
		}

		if meta, ok := metadataMap[name]; ok {
			info.Path = meta.Path
			info.Purpose = meta.Purpose
			info.CreatedAt = meta.CreatedAt
			info.UpdatedAt = meta.UpdatedAt
		}

		// Default workspace path is the repo root
		if name == "default" && info.Path == "" {
			info.Path = root
		}

		// Check if current
		currentRoot, _ := ws.Root()
		if info.Path == currentRoot {
			info.IsCurrent = true
		}

		display.PrintShowDetail(os.Stdout, info)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
