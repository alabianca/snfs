package services

import (
	"encoding/json"
	"io"
)

func decode(body io.Reader, d interface{}) error {
	return json.NewDecoder(body).Decode(d)
}
