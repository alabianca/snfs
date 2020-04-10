package watchdog

import (
	"encoding/json"
	"errors"
	"github.com/alabianca/snfs/snfsd"
	"io"
	"log"
	"strconv"
	"sync"
)


type watchdog struct {
	exit chan struct{}
	closed bool
	mtx sync.RWMutex
	pubsub snfsd.PubSub
	writer io.Writer
}

func New(pubsub snfsd.PubSub, logDigester io.Writer) snfsd.Watchdog {
	return &watchdog{
		exit:   make(chan struct{}),
		closed: false,
		mtx:    sync.RWMutex{},
		pubsub: pubsub,
		writer: logDigester,
	}
}

func (m *watchdog) Watch() {
	// subscribe to new nodes being added
	pcr := m.startProcReq()
	// accumulate new nodes and store them
	cp := m.collect(m.exit, pcr)
	// start newly added nodes as child processes
	ccp := m.startProcesses(cp)
	// monitor child processes
	mon := m.monitor(ccp)
	// kill processes that end
	m.kill(mon)
}

func (m *watchdog) Close() {
	log.Println("Closing Watchdog")
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if !m.closed {
		m.closed = true
		close(m.exit)
	}
}

func (m *watchdog) Write(p []byte) (int, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.writer.Write(p)
}

func (m *watchdog) startProcReq() chan startProcessRequest {
	req := m.pubsub.Subscribe(snfsd.TopicAddNode)
	out := make(chan startProcessRequest)

	go func() {
		for data := range req {
			var conf snfsd.NodeConfiguration
			if err := json.Unmarshal(data, &conf); err == nil {
				opt := make(options)
				opt["-cport"] = strconv.Itoa(conf.Cport)
				opt["-dport"] = strconv.Itoa(conf.Dport)
				out <- startProcessRequest{process:Process(conf.Name, opt, m), errc: make(chan error, 1)}
			}
		}
	}()

	return out
}

// collect collect's processes that need to be started
func (m *watchdog) collect(done chan struct{}, queue chan startProcessRequest) chan *process {
	out := make(chan *process)

	go func() {
		defer func() {
			log.Printf("Closing Collect")
			close(out)
		}()
		nodes := getNodes()

		for {
			select {
			case <-done:
				return

			// new process that started up
			case req := <- queue:
				if !nodes.push(req.process) {
					req.errc <- errors.New("Already exists")
					break
				}

				out <- req.process
				close(req.errc)


			}
		}
	}()

	return out
}

func (m *watchdog) startProcesses(p chan *process) chan chan *process {
	running := make(chan chan *process)

	go func() {
		defer func() {
			log.Printf("Closing start processes")
			close(running)
		}()
		for proc := range p {
			go func(proc *process, n *nodes) {
				n.update(proc.name, toggleRunning(true))
				// run the process
				running <- proc.run()

			}(proc, getNodes())
		}
	}()

	return running
}

func (m *watchdog) monitor(in chan chan *process) chan *process {
	out := make(chan *process)
	go func() {
		defer func() {
			log.Printf("Closing Monitor")
			close(out)
		}()
		// monitor the running processes
		for runningProc := range in {
			go func(c chan *process) {
				p := <- runningProc
				// the process exited
				out <- p
			}(runningProc)
		}
	}()

	return out
}

func (m *watchdog) kill(in chan *process) {
	n := getNodes()
	for p := range in {
		log.Printf("Update %d\n", p.id())
		n.update(p.name, toggleRunning(false))
		log.Printf("Process %d killed (%s)\n", p.id(), p.name)
	}

	// make sure any nodes that may still be running must be stopped to prevent run away processes
	log.Printf("Checking for runaway processes\n")
	n.foreach(func(p *process, i int) {
		if p.running {
			log.Printf("Killing Potential Run away process %d\n", p.id())
			p.kill()
			log.Printf("Killed\n")
			p.running = false
		} else {
			log.Printf("Process %d already exited\n", p.id())
		}
	})
}

// utils
func toggleRunning(running bool) func(err error, p *process) {
	return func(err error, p *process) {
		if err == nil {
			p.running = running
		}
	}
}

