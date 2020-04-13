package log

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/alabianca/snfs/snfsd"
	"io"
	"log"
	"os"
	"path"
	"time"
)

type LogEvent struct {
	Tag       string      `json:"tag"`
	Time      time.Time   `json:"time"`
	EventName string      `json:"event"`
	Data      interface{} `json:"data"`
}

type readResult struct {
	event []byte
	err error
}

type eventLogger struct {
	file *os.File
	logger *log.Logger
}

func (e *eventLogger) Println(v ...interface{}) {
	e.logger.Println(v...)
}

func (e *eventLogger) Close() error {
	return e.file.Close()
}

type Digester struct {
	NodeService snfsd.NodeService
	Writer      io.Writer
	Reader      io.Reader
	Root        string
	Logs        map[string]snfsd.Logger
	exit        chan struct{}
}

func New(ns snfsd.NodeService) snfsd.LogDigest {
	return &Digester{
		Logs: make(map[string]snfsd.Logger),
	}
}

func (d *Digester) Process() error {
	r, w := io.Pipe()
	d.Reader = r
	d.Writer = w
	d.exit = make(chan struct{})

	defer w.Close()
	defer d.close()


	for ev := range d.read(d.exit) {
		logEv, err := d.decodeEvent(ev)
		if err != nil {
			continue
		}

		if err := d.processEvent(&logEv); err != nil {
			log.Printf("Error processing event %s\n", err)
		}
	}

	return nil
}

func (d *Digester) Close() {
	close(d.exit)
}

func (d *Digester) close() {
	log.Println("Closing Open Log Files")
	for _, v := range d.Logs {
		v.Close()
	}
}

// read reads events line by line sending them to 'out'
// reading is done async and events are buffered before sending them to out
// once 'exit' is closed we flush out all buffered events
func (d *Digester) read(exit chan struct{}) <- chan []byte {
	out := make(chan []byte)
	go func() {
		defer close(out)

		var readDone chan readResult
		var next time.Time
		var drain bool

		buf := make([][]byte, 0)
		maxBuf := 10
		reader := bufio.NewReader(d.Reader)

		for {
			var readDelay time.Duration
			var startRead <-chan time.Time
			if now := time.Now(); next.After(now) {
				readDelay = next.Sub(now) // starts out as 0;
			}

			if !drain && readDone == nil && len(buf) < maxBuf {
				startRead = time.After(readDelay)
			}

			var fanout chan []byte
			var nextEvent []byte
			if len(buf) > 0 {
				fanout = out
				nextEvent = buf[0]
			}

			if drain && len(buf) == 0 {
				break
			}

			select {
			case <-exit:
				drain = true
			case <-startRead:
				readDone = make(chan readResult, 1)

				go func() {
					ln, _, err := reader.ReadLine()
					readDone <- readResult{ln,err}
				}()

			case result := <- readDone:
				readDone = nil
				next = time.Now().Add(time.Nanosecond * 1)
				if result.err == nil {
					buf = append(buf, result.event)
				}

			case fanout <- nextEvent:
				buf = buf[1:]
			}
		}
	}()

	return out
}

func (d *Digester) decodeEvent(event []byte) (LogEvent, error) {
	var le LogEvent
	err := json.Unmarshal(event, &le)
	return le, err
}

func (d *Digester) processEvent(event *LogEvent) (error) {
	var ok bool
	var nc snfsd.NodeConfiguration
	switch event.EventName {
	case nodeStarted:
		nc, ok = event.Data.(snfsd.NodeConfiguration)
		if ok {
			d.toggleNode(nc, true)
		}
	case nodeExited:
		nc, ok = event.Data.(snfsd.NodeConfiguration)
		if ok {
			d.toggleNode(nc, false)
		}
	}

	if !ok {
		return errors.New("unknown event")
	}

	if nc.NodeId != "" {
		logger, err := d.getLoggerFor(nc.NodeId)
		if err != nil {
			return err
		}

		return d.writeLogEvent(event, logger)
	}

	return errors.New("unkown event")
}

func (d *Digester) getLoggerFor(nodeId string) (snfsd.Logger, error) {
	if v, ok := d.Logs[nodeId]; ok {
		return v, nil
	}

	// else create the log file for the nodeId
	f, err := os.OpenFile(path.Join(d.Root, nodeId + ".log"), os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	logger := &eventLogger{
		file:   f,
		logger: log.New(f, nodeId, log.LstdFlags),
	}

	d.Logs[nodeId] = logger

	return logger, nil

}

func (d *Digester) toggleNode(nc snfsd.NodeConfiguration, up bool) error {
	if up {
		nc.Started = time.Now().Unix()
	} else {
		nc.Started = 0
	}

	return d.NodeService.Update(nc)
}

func (d *Digester) writeLogEvent(event *LogEvent, logger snfsd.Logger) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	logger.Println(string(b))
	return nil
}

func (d *Digester) Write(p []byte) (int, error) {
	return d.Writer.Write(p)
}
