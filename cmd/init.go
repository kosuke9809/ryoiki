package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kosuke9809/ryoiki/internal/shell"
)

var initCmd = &cobra.Command{
	Use:   "init <shell>",
	Short: "Set up shell integration",
	Long:  "Set up shell integration for ryoiki switch.\n\nBy default, appends the eval line to your shell config file.\nUse --print to output the script to stdout instead.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shellName := args[0]
		printOnly, _ := cmd.Flags().GetBool("print")

		if printOnly {
			script, err := shell.InitScript(shellName)
			if err != nil {
				return err
			}
			fmt.Print(script)
			return nil
		}

		// Default: auto-write to config file
		configPath, err := shell.ConfigFilePath(shellName)
		if err != nil {
			return err
		}

		already, err := shell.WriteToConfig(shellName)
		if err != nil {
			return err
		}

		if already {
			fmt.Println("Shell integration is already configured.")
			return nil
		}

		fmt.Printf("Added shell integration to %s\n", configPath)
		fmt.Println("Restart your shell or run: source " + configPath)
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("print", false, "Print the shell script to stdout instead of writing to config file")
	rootCmd.AddCommand(initCmd)
}
