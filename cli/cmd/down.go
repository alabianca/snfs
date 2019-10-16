package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Kill node",
	Long:  `Stop receiving connections from peers`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
