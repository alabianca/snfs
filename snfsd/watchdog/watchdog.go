package watchdog

import (
	"encoding/json"
	"errors"
	"github.com/alabianca/snfs/snfsd"
	"log"
	"strconv"
	"sync"
)


type watchdog struct {
	exit chan struct{}
	closed bool
	mtx sync.RWMutex
	pubsub snfsd.PubSub
}

func New(pubsub snfsd.PubSub) snfsd.Watchdog {
	return &watchdog{
		exit:   make(chan struct{}),
		closed: false,
		mtx:    sync.RWMutex{},
		pubsub: pubsub,
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
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if !m.closed {
		m.closed = true
		close(m.exit)
	}
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
				opt["-fport"] = strconv.Itoa(conf.Fport)
				out <- startProcessRequest{process:Process(conf.Name, opt), errc: make(chan error, 1)}
			}
		}
	}()

	return out
}

// collect collect's processes that need to be started
func (m *watchdog) collect(done chan struct{}, queue chan startProcessRequest) chan *process {
	out := make(chan *process)

	go func() {
		defer close(out)
		nodes := getNodes()

		for {
			select {
			case <-done:
				return

			// new process that started up
			case req := <- queue:
				if !nodes.push(req.process) {
					req.errc <- errors.New("Already exists")
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
		defer close(running)
		for proc := range p {
			go func(proc *process, n *nodes) {
				n.update(proc.name, toggleRunning(true))

				running <- proc.run()

			}(proc, getNodes())
		}
	}()

	return running
}

func (m *watchdog) monitor(in chan chan *process) chan *process {
	out := make(chan *process)
	go func() {
		defer close(out)

		for runningProc := range in {
			go func(c chan *process) {
				p := <- runningProc
				out <- p
			}(runningProc)
		}
	}()

	return out
}

func (m *watchdog) kill(in chan *process) {
	n := getNodes()
	for p := range in {
		log.Printf("Process %d killed (%s)\n", p.id(), p.name)
		n.update(p.name, toggleRunning(false))
	}

	// make sure any nodes that may still be running must be stopped
	n.foreach(func(p *process, i int) {
		if p.running {
			p.kill()
			p.running = false
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

