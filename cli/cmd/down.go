package cmd

import (
	"fmt"
	"github.com/alabianca/snfs/cli"

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
		mdns := cli.NewMdnsService()
		status, err := mdns.Unregister()

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(status)
	},
}
