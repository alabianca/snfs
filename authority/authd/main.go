package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type CSR struct {
	CommonName string `json:"common_name"`
	TTL        string `json:"ttl"`
}

type CSRResponse struct {
	RequestID     string          `json:"request_id"`
	LeaseID       string          `json:"lease_id"`
	Renewable     bool            `json:"renewable"`
	LeaseDuration int32           `json:"lease_duration"`
	Data          CSRResponseData `json:"data"`
}

type CSRResponseData struct {
	CAChain        []string `json:"ca_chain"`
	Certificate    string   `json:"certificate"`
	IssuingCA      string   `json:"issuing_ca"`
	PrivateKey     string   `json:"private_key"`
	PrivateKeyType string   `json:"private_key_type"`
	SerialNumber   string   `json:"serial_number"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Could not load environment file")
	}

	vaultToken := os.Getenv("VAULT_DEV_TOKEN")
	fmt.Println(vaultToken)

	client := &http.Client{}
	csr := &CSR{
		CommonName: "lucky-goat.snfs.com",
		TTL:        "24h",
	}

	bts, err := json.Marshal(csr)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(bts)
	fmt.Println("Sending CSR")
	fmt.Println(buf.String())

	req, err := http.NewRequest("POST", "http://127.0.0.1:8200/v1/pki_int/issue/snfs-dot-com", buf)
	if err != nil {
		panic(err)
	}

	req.Header.Add("X-Vault-Token", vaultToken)
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var csrResp CSRResponse
	if err := json.Unmarshal(body, &csrResp); err != nil {
		panic(err)
	}

	fmt.Println(csrResp.Data.Certificate)
	fmt.Println(csrResp.Data.PrivateKey)
	fmt.Println(csrResp.Data.IssuingCA)

}
