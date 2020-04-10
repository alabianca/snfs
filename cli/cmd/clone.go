package cmd

import (
	"fmt"
	"github.com/alabianca/snfs/cli"
	"log"
	"os"

	"github.com/alabianca/spin"

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
			log.Fatal("Please provide file hash to download")
		}

		runClone(args[0])
	},
}

func runClone(fileHash string) {
	spinner := spin.NewSpinner(spin.Dots2, os.Stdout)
	errChan := make(chan error)
	successChan := make(chan bool)

	go initSpinnerWithText(spinner, fmt.Sprintf("Downloading -> %s", fileHash))
	go clone(fileHash, successChan, errChan)

	select {
	case err := <-errChan:
		spinner.Stop()
		fmt.Printf("[Error] %s\n", err)
	case <-successChan:
		spinner.Stop()
		fmt.Println("Content downloaded")
	}
}

func clone(fileHash string, success chan bool, errc chan error) {
	storageService := cli.NewStroageService()
	if err := storageService.Download(fileHash); err != nil {
		errc <- err
		return
	}

	success <- true
}
