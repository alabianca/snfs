package cmd

import (
	"fmt"
	"github.com/alabianca/snfs/cli"
	"github.com/spf13/cobra"
	"log"
	"os"
	"text/tabwriter"
)

func init() {
	kadnetCmd.AddCommand(kadnetStatusCmd)
}

var kadnetStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of your DHT.",
	Long:  `Prints a snapshot of the DHT`,
	Run: func(cmd *cobra.Command, args []string) {
		kad := cli.NewKadnetService()
		entries, err := kad.GetStatus()
		if err != nil {
			log.Fatalf("Error %s\n", err)
		}


		fmt.Printf("Found %d Contacts\n", len(entries))
		fmt.Println()
		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
		fmt.Fprintln(writer, "Bucket Index\tID\tAddress\tPort\t")
		for _, i := range entries {
			fmt.Fprintf(writer, "%d\t%s\t%s\t%d\t\n", i.BucketIndex, i.Contact.ID, i.Contact.IP, i.Contact.Port)
		}
		writer.Flush()
		fmt.Println()
	},
}
