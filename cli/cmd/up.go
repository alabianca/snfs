package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

type subscriptionRequest struct {
	Instance string `json:"instance"`
}

var upCmd = &cobra.Command{
	Use:   "up [instance] [port]",
	Short: "Register node",
	Long:  `Register your node in the local network. Accept connections from peers`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Please provide instance")
		}

		data, _ := marshalSubscriptionRequest(args[0])
		body := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", "http://localhost:4200/api/v1/mdns/subscribe", body)
		if err != nil {
			log.Fatal(err)
		}

		client := http.Client{}
		client.Do(req)

	},
}

func marshalSubscriptionRequest(instance string) ([]byte, error) {
	req := subscriptionRequest{
		Instance: instance,
	}

	return json.Marshal(&req)
}
