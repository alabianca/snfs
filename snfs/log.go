package snfs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type LogEvent struct {
	Tag       string      `json:"tag"`
	Time      time.Time   `json:"time"`
	EventName string      `json:"event"`
	Data      interface{} `json:"data"`
}

type Logger struct {
	stdout   io.Writer
	newEvent chan LogEvent
	exit     chan struct{}
}

var EventLogger *Logger

var defLogger *Logger
var createOnce sync.Once

func defaultLogger() *Logger {
	createOnce.Do(func() {
		defLogger = NewLogger(os.Stdout)
	})

	return defLogger
}

var mtx sync.Mutex

func GetLogger() *Logger {
	mtx.Lock()
	defer mtx.Unlock()
	if EventLogger == nil {
		return defaultLogger()
	}

	return EventLogger
}

func NewLogger(stdout io.Writer) *Logger {
	return &Logger{
		stdout:   stdout,
		newEvent: make(chan LogEvent),
		exit:     make(chan struct{}),
	}
}

func (l *Logger) Start() {
	go func() {
		for {
			select {
			case <-l.exit:
				return
			case event := <-l.newEvent:
				b, _ := json.Marshal(event)

				fmt.Fprintln(l.stdout, string(b))
			}
		}
	}()
}

func (l *Logger) Exit() {
	close(l.exit)
}

func (l *Logger) NewEvent(tag, name string, ev interface{}) {
	l.newEvent <- LogEvent{
		Tag:tag,
		EventName: name,
		Time: time.Now(),
		Data: ev,
	}
}
