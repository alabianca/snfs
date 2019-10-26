package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/alabianca/snfs/cli/services"

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
		var fname string
		if len(args) < 1 {
			log.Fatal("Please provide file name")
		}
		if len(args) == 2 {
			fname = args[1]
		}

		storage := services.NewStroageService()
		uploadCntx := args[0]

		finfo, err := os.Stat(uploadCntx)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(finfo.IsDir())
			fmt.Println(finfo.Name())
			fmt.Println(finfo.Size())
			storage.Upload(fname, uploadCntx)
		}
	},
}
