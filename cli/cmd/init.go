package cmd

import (
	"fmt"
	"github.com/alabianca/snfs/cli"
	"github.com/spf13/cobra"
	"log"
	"strconv"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [dport] [cport] [name]",
	Args: cobra.MinimumNArgs(3),
	Short: "Initializes a new node",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cli.CreateSnfsDir(); err != nil {
			fmt.Println("There was an error trying to create the SNFS directory")
			return
		}

		dport, err1 := strconv.ParseInt(args[0], 10, 16)
		cport, err2 := strconv.ParseInt(args[1], 10, 16)
		name := args[2]
		if err1 != nil || err2 != nil {
			fmt.Println("Could not parse cport or dport")
			return
		}

		node := cli.NewNodeService()
		nc, err := node.Create(int(cport), int(dport), name)
		if err != nil {
			fmt.Printf("Error %s\n", err)
			return
		}

		log.Println(nc)

	},
}
