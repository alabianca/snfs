package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap [port] [ip] [id]",
	Short: "Bootstrap your node and join the network",
	Args:  cobra.MinimumNArgs(3),
	Long:  `Bootstrap your node and join the network by contacting a bootstrap node (id) at port,ip `,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
