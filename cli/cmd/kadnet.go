package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(kadnetCmd)
}

var kadnetCmd = &cobra.Command{
	Use:   "kadnet",
	Short: "Interact with the DHT",
	Long:  `Commands to interact with the underlying kademlia dht`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}