package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List nodes",
	Long:  `List all sharing nodes in the local network. You can clone content from these nodes`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(time.Now())
	},
}
