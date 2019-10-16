package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Register node",
	Long:  `Register your node in the local network. Accept connections from peers`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
