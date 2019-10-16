package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cloneCmd)
}

var cloneCmd = &cobra.Command{
	Use:   "clone [node to clone]",
	Short: "Clone content",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Clone the contents of a particular node into your current working directory`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(time.Now())
	},
}
