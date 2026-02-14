package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var forgetCmd = &cobra.Command{
	Use:   "forget <name>",
	Short: "Remove a workspace",
	Long:  "Remove a workspace from jj and optionally delete its directory.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")
		removeDir, _ := cmd.Flags().GetBool("remove-dir")

		if name == "default" {
			return fmt.Errorf("cannot forget the default workspace")
		}

		ws, store, _, err := initServices()
		if err != nil {
			return err
		}

		// Confirmation prompt unless --force
		if !force {
			msg := fmt.Sprintf("Forget workspace %q?", name)
			if removeDir {
				msg = fmt.Sprintf("Forget workspace %q and remove its directory?", name)
			}
			fmt.Printf("%s [y/N] ", msg)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Get the workspace path before forgetting (for directory removal)
		var wsPath string
		if removeDir {
			wsPath, _ = store.GetPath(name)
		}

		// Forget workspace in jj
		if err := ws.Forget(name); err != nil {
			return err
		}

		// Remove metadata
		if err := store.Remove(name); err != nil {
			return fmt.Errorf("workspace forgotten but failed to remove metadata: %w", err)
		}

		fmt.Printf("Forgot workspace %q\n", name)

		// Remove directory if requested
		if removeDir && wsPath != "" {
			if err := os.RemoveAll(wsPath); err != nil {
				return fmt.Errorf("workspace forgotten but failed to remove directory %q: %w", wsPath, err)
			}
			fmt.Printf("Removed directory: %s\n", wsPath)
		}

		return nil
	},
}

func init() {
	forgetCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
	forgetCmd.Flags().Bool("remove-dir", false, "Also remove workspace directory")
	rootCmd.AddCommand(forgetCmd)
}
