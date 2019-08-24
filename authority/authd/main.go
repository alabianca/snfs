package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
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
	vaultURL := os.Getenv("VAULT_URL")

	http.Handle("/certificate", certificate(vaultToken, vaultURL))

	if err := http.ListenAndServe(getListenAddr(), nil); err != nil {
		log.Println(err)
	}

	// client := &http.Client{}
	// csr := &CSR{
	// 	CommonName: "lucky-goat.snfs.com",
	// 	TTL:        "24h",
	// }

	// bts, err := json.Marshal(csr)
	// if err != nil {
	// 	panic(err)
	// }

	// buf := bytes.NewBuffer(bts)
	// fmt.Println("Sending CSR")
	// fmt.Println(buf.String())

	// req, err := http.NewRequest("POST", vaultURL, buf)
	// if err != nil {
	// 	panic(err)
	// }

	// req.Header.Add("X-Vault-Token", vaultToken)
	// resp, err := client.Do(req)

	// if err != nil {
	// 	panic(err)
	// }

	// defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	panic(err)
	// }

	// var csrResp CSRResponse
	// if err := json.Unmarshal(body, &csrResp); err != nil {
	// 	panic(err)
	// }

	// fmt.Println(csrResp.Data.Certificate)
	// fmt.Println(csrResp.Data.PrivateKey)
	// fmt.Println(csrResp.Data.IssuingCA)

}

func getListenAddr() string {
	port := os.Getenv("PORT")
	return ":" + port
}

func certificate(vaultToken, vaultURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, r *http.Request) {
		csr, err := parseRequestBody(r)
		if err != nil {
			res.WriteHeader(400)
			res.Write([]byte(err.Error()))
			return
		}

		log.Printf("Getting Certificate for %s\n", csr.CommonName)
		client := &http.Client{}
		vReq, err := vaultRequest(csr, vaultURL, vaultToken)
		if err != nil {
			res.WriteHeader(400)
			res.Write([]byte(err.Error()))
			return
		}

		resp, err := client.Do(vReq)

		if err != nil {
			res.WriteHeader(400)
			res.Write([]byte(err.Error()))
			return
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			res.WriteHeader(400)
			res.Write([]byte(err.Error()))
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(body)

	}
}

func requestBodyDecoder(request *http.Request) (*json.Decoder, error) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body))), nil

}

func parseRequestBody(request *http.Request) (CSR, error) {
	decoder, err := requestBodyDecoder(request)
	if err != nil {
		return CSR{}, err
	}

	var csr CSR
	decoder.Decode(&csr)
	return csr, err

}

func vaultRequest(csr CSR, url, token string) (*http.Request, error) {
	bts, err := json.Marshal(csr)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(bts)

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Vault-Token", token)

	return req, nil
}
