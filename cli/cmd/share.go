package cmd

import (
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"github.com/alabianca/snfs/cli"
	"log"
	"os"
	"time"

	"github.com/alabianca/spin"

	"github.com/alabianca/snfs/util"

	"github.com/spf13/cobra"
)

var tag string

const (
	GB = 1000000000 // 1 Gigabytes
	MB = 1000000    // 1 Megabyte
	KB = 1000       // 1 Byte
)

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
	res := make(chan cli.StoreResult)

	go initSpinnerWithText(spinner, fmt.Sprintf("Sharing -> %s", uploadCntx))
	go upload(errChan, res, fname, uploadCntx)

	select {
	case err := <-errChan:
		spinner.Stop()
		time.Sleep(time.Millisecond * 200)
		fmt.Printf("[Error]: %s\n", err)
		return err

	case hash := <-res:
		spinner.Stop()
		time.Sleep(time.Millisecond * 200)
		fmt.Println()
		printUploadResult(hash)
	}

	return nil

}

func upload(errChan chan error, chanSuccess chan cli.StoreResult, fname, uploadCntx string) {
	storage := cli.NewStroageService()

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
	hash := sha1.New()
	gzw := gzip.NewWriter(hash)
	if _, err := util.WriteTarball(gzw, uploadCntxt); err != nil {
		return nil, err
	}

	gzw.Close()

	return hash.Sum(nil), nil

}

func printUploadResult(res cli.StoreResult) {
	fmt.Println()
	fmt.Println(White("Success"))
	fmt.Println()
	fmt.Printf("%s           %s\n", White("Hash:"), Green(res.Hash))
	fmt.Printf("%s  %s (Uncompressed)\n", White("Bytes Written:"), Green(formatBytes(res.BytesWritten)))
	fmt.Printf("%s           %s\n", White("Took:"), Green(res.Took))
	fmt.Println()
	fmt.Printf("To get the content %s %s\n", White("snfs clone"), White(res.Hash))
	fmt.Println()
}

func formatBytes(numBytes int64) string {
	var res float32

	if numBytes >= GB {
		res = float32(numBytes) / float32(GB)
		return fmt.Sprintf("%.2f GB", res)
	}

	if numBytes >= MB {
		res = float32(numBytes) / float32(MB)
		return fmt.Sprintf("%.2f MB", res)
	}

	if numBytes >= KB {
		res = float32(numBytes) / float32(KB)
		return fmt.Sprintf("%.2f KB", res)
	}

	return fmt.Sprintf("%d Bytes", numBytes)
}
