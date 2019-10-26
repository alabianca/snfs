package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(shareCmd)
}

var shareCmd = &cobra.Command{
	Use:   "share [share context]",
	Short: "Share content",
	Args:  cobra.MinimumNArgs(1),
	Long:  `share the given context with peers in the local network`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(time.Now())

		finfo, err := os.Stat(args[0])
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(finfo.IsDir())
			fmt.Println(finfo.Name())
		}
	},
}
