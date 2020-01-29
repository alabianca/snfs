package cmd

import (
	"log"
	"strconv"

	"github.com/alabianca/snfs/cli/services"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [port] [ip]",
	Short: "Bootstrap your node and join the network",
	Args:  cobra.MinimumNArgs(2),
	Long:  `Bootstrap your node and join the network by contacting a bootstrap node (id) at port,ip `,
	Run: func(cmd *cobra.Command, args []string) {
		bs := services.NewKadnetService()

		port, err := strconv.ParseInt(args[0], 10, 16)
		if err != nil {
			log.Fatal(err)
		}

		bs.Boostrap(int(port), args[1])
	},
}
