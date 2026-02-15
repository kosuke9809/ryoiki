package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/config"
	"github.com/kosuke9809/ryoiki/internal/jj"
)

var rootCmd = &cobra.Command{
	Use:   "ryoiki",
	Short: "jj workspace management tool",
	Long:  "ryoiki - A CLI tool specialized for jj (Jujutsu) workspace management.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// initServices creates the workspace service and metadata store.
// Returns the workspace service, metadata store, and the repo root path.
func initServices() (*jj.WorkspaceService, *config.MetadataStore, string, error) {
	executor := &jj.DefaultExecutor{}
	ws := jj.NewWorkspaceService(executor)
	root, err := ws.Root()
	if err != nil {
		return nil, nil, "", fmt.Errorf("not in a jj repository: %w", err)
	}
	store := config.NewMetadataStore(root)
	return ws, store, root, nil
}
