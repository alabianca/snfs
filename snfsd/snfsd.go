package snfsd

import (
	//"context"
	//"github.com/alabianca/snfs/snfsd/api"
	//"github.com/alabianca/snfs/snfsd/watchdog"
	//"log"
	//"net/http"
	//"os"
	//"os/signal"
	//"syscall"
	//"time"
	"encoding/json"
	"io"
	"net/http"
)

type Handler http.Handler

func Decode(reader io.Reader, out interface{}) error {
	decoder := json.NewDecoder(reader)
	return decoder.Decode(out)
}

func Encode(writer io.Writer, in interface{}) error {
	encoder := json.NewEncoder(writer)
	return encoder.Encode(in)
}
