package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ry",
	Short: "jj workspace management tool",
	Long:  "ryoiki - A CLI tool specialized for jj (Jujutsu) workspace management.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
