package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/alabianca/snfs/cli/services"

	"github.com/alabianca/spin"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List nodes",
	Long:  `List all sharing nodes in the local network. You can clone content from these nodes`,
	Run: func(cmd *cobra.Command, args []string) {

		runLs()
	},
}

func runLs() {
	sig := make(chan os.Signal, 1)
	browser := make(chan []services.Node)
	errChan := make(chan error)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	spinner := spin.NewSpinner(spin.Dots2, os.Stdout)

	go initSpinnerWithText(spinner, "Browsing Local Network")
	go browse(browser, errChan)

	select {
	case <-sig:
		spinner.Stop()
	case instances := <-browser:
		spinner.Stop()
		printBrowseResults(instances)
	case err := <-errChan:
		spinner.Stop()
		printError(err)
	}

}

func printError(err error) {
	fmt.Printf("[Error]: %s\n", err)
}

func printBrowseResults(instances []services.Node) {
	fmt.Printf("Found %d Result(s)\n", len(instances))
	fmt.Println()
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(writer, "Instance\tPort\tAddress\tID\t")
	for _, i := range instances {
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\t\n", i.InstanceName, i.Port, i.Address, i.ID)
	}
	writer.Flush()
	fmt.Println()
}

func initSpinnerWithText(spinner spin.Spinner, text string) {
	fmt.Printf("  %s", text)
	spinner.Start()
}

func browse(done chan []services.Node, errChan chan error) {
	mdns := services.NewMdnsService()
	instances, err := mdns.Browse()
	if err != nil {
		errChan <- err
		return
	}

	done <- instances
}
