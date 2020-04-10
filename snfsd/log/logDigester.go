package log

import (
	"github.com/alabianca/snfs/snfsd"
	"io"
)

type Digester struct {
	NodeService snfsd.NodeService
	writer io.Writer
	reader io.Reader
}

func New() snfsd.LogDigest {
	r, w := io.Pipe()
	return &Digester{
		writer: w,
		reader: r,
	}
}

func (d *Digester) Process() error {
	return nil
}

func (d *Digester) read() {

}

func (d *Digester) processEvent(event string) {

}

func (d *Digester) Write(p []byte) (int, error) {
	return d.writer.Write(p)
}
