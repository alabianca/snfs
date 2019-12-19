package services

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
)

func decode(body io.Reader, d interface{}) error {
	return json.NewDecoder(body).Decode(d)
}

func getPort() int64 {
	port := os.Getenv("PORT")
	// default to 4200
	p, err := strconv.ParseInt(port, 10, 16)
	if err != nil {
		return 4200
	}

	return p

}

func getBaseURL() string {
	p := strconv.Itoa(int(getPort()))
	return "http://localhost:" + p + "/api/"

}
