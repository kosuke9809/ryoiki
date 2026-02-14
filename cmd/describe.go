package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Set workspace purpose",
	Long:  "Set or update the purpose/description for a workspace.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		purpose, _ := cmd.Flags().GetString("purpose")

		if purpose == "" {
			return fmt.Errorf("purpose is required; use --purpose or -p to specify")
		}

		_, store, _, err := initServices()
		if err != nil {
			return err
		}

		if err := store.SetPurpose(name, purpose); err != nil {
			return err
		}

		fmt.Printf("Updated workspace %q purpose: %s\n", name, purpose)
		return nil
	},
}

func init() {
	describeCmd.Flags().StringP("purpose", "p", "", "Purpose description")
	rootCmd.AddCommand(describeCmd)
}
