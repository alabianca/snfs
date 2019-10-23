package cmd

import (
	"fmt"
	"log"

	"github.com/alabianca/snfs/cli/services"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up [instance]",
	Short: "Register node",
	Long:  `Register your node in the local network. Accept connections from peers`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Please provide instance")
		}

		mdns := services.NewMdnsService()
		status, err := mdns.Register(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(status)

	},
}
