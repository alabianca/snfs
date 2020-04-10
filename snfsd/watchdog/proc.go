package watchdog

import (
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type process struct {
	_id int
	running bool
	cport string
	dport string
	name string
	command *exec.Cmd
}

func (p *process) kill() error {
	return p.command.Process.Kill()
}

func (p *process) id() int {
	if p.command == nil || p.command.Process == nil {
		return p._id
	}

	return p.command.Process.Pid
}

func (p *process) run() chan *process {
	out := make(chan *process)

	go func() {
		p.command.Run()
		out <- p
	}()

	return out
}
func Process(name string, opt options, writer io.Writer) *process {
	log.Printf("snfsd_node %v\n", opt)
	cmd := exec.Command("snfsd_node", opt.Args()...)
	cmd.Stdout = writer

	return &process{
		cport:   opt["-cport"],
		dport:   opt["-dport"],
		name:    name,
		command: cmd,
	}
}

type nodes struct {
	processes map[string]*process
	mtx *sync.Mutex
}

func (n *nodes) push(p *process) bool {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	if _, ok := n.processes[p.name]; ok {
		return false
	}
	n.processes[p.name] = p

	return true
}

func (n *nodes) update(name string, up func(err error, p *process)) {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	p, ok := n.processes[name]
	if !ok {
		up(errors.New("Process Not Found"), nil)
	} else {
		up(nil, p)
	}
}

func (n *nodes) foreach(up func(p *process, i int)) {
	n.mtx.Lock()
	defer n.mtx.Unlock()

	var index int
	for _, p := range n.processes {
		up(p, index)
		index++
	}
}

type options map[string]string
func (o options) String() string {
	out := ""
	for k, v := range o {
		out = out + k + " " + v + " "
	}

	return out
}
func (o options) Args() []string {
	return strings.Split(o.String(), " ")
}


type startProcessRequest struct {
	process *process
	errc chan error
}

var once sync.Once
var storedNodes *nodes
func getNodes() *nodes {
	once.Do(func() {
		storedNodes = &nodes{
			processes: make(map[string] *process),
			mtx:       new(sync.Mutex),
		}
	})

	return storedNodes
}
