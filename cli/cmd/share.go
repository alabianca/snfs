package cmd

import (
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"log"
	"os"

	"github.com/alabianca/spin"

	"github.com/alabianca/snfs/cli/services"
	"github.com/alabianca/snfs/util"

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

		uploadCntx := args[0]
		fname = tag

		runShare(uploadCntx, fname)
	},
}

func runShare(uploadCntx, fname string) error {
	_, err := os.Stat(uploadCntx)
	if err != nil {
		return err
	}

	spinner := spin.NewSpinner(spin.Dots2, os.Stdout)
	errChan := make(chan error)
	resultHash := make(chan string)

	go initSpinnerWithText(spinner, fmt.Sprintf("Sharing -> %s", uploadCntx))
	go upload(errChan, resultHash, fname, uploadCntx)

	select {
	case err := <-errChan:
		spinner.Stop()
		fmt.Printf("[Error]: %s\n", err)
		return err

	case hash := <-resultHash:
		spinner.Stop()
		printUploadResult(hash)
	}

	return nil

}

func upload(errChan chan error, chanSuccess chan string, fname, uploadCntx string) {
	storage := services.NewStroageService()

	if fname == "" {
		if bts, err := hashContents(uploadCntx); err != nil {
			errChan <- err
			return
		} else {
			fname = fmt.Sprintf("%x", bts)
		}
	}

	if resultHash, err := storage.Upload(fname, uploadCntx); err != nil {
		errChan <- err
	} else {
		chanSuccess <- resultHash
	}

}

func hashContents(uploadCntxt string) ([]byte, error) {
	hash := md5.New()
	gzw := gzip.NewWriter(hash)
	if err := util.WriteTarball(gzw, uploadCntxt); err != nil {
		return nil, err
	}

	gzw.Close()

	return hash.Sum(nil), nil

}

func printUploadResult(hash string) {
	fmt.Printf("Uploaded -> %s\n", hash)
}
