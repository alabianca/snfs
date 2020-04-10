package cli

import (
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"io"
	"os"
	"path"
	"strconv"
)

func decode(body io.Reader, d interface{}) error {
	return json.NewDecoder(body).Decode(d)
}

func getPort() int64 {
	port := os.Getenv("SNFSD_PORT")
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

func getHomeDir() (string, error) {
	return homedir.Dir()
}

// create the SnfsDir if it does not yet exist
// returns nil if it creates it successfully or the directory already exists
func CreateSnfsDir() error {
	home, err := getHomeDir()
	if err != nil {
		return err
	}

	dir := path.Join(home, ".snfs")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		// directory already exists
		return nil
	}

	return os.Mkdir(path.Join(home, ".snfs"), 0755)
}