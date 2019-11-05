package cmd

import (
	"log"

	"github.com/alabianca/snfs/cli/services"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cloneCmd)
}

var cloneCmd = &cobra.Command{
	Use:   "clone [hash]",
	Short: "Clone content",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Clone the contents of a particular node into your current working directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Please provide file to download")
		}

		storageService := services.NewStroageService()

		storageService.Download(args[0])
	},
}
