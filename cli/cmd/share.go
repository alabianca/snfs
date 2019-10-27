package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/alabianca/snfs/cli/services"

	"github.com/spf13/cobra"
)

var tag string

func init() {
	rootCmd.AddCommand(shareCmd)
	shareCmd.Flags().StringVarP(&tag, "tag", "t", "", "Override default file name that is created")
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

		storage := services.NewStroageService()
		uploadCntx := args[0]
		fname = tag

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
