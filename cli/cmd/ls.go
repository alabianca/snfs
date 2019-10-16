package cmd

import (
	"os"
	"time"

	"github.com/alabianca/spin"
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

		spinner := spin.NewSpinner(spin.Dots, os.Stdout)

		go func() {
			spinner.Start()
		}()

		<-time.After(time.Second * 5)
		spinner.Stop()
	},
}
