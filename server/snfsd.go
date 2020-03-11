package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type runningProc map[int]*exec.Cmd
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


func main() {
	cports := []string{"4200", "4300", "4400"}
	dports := []string{"5050", "5060", "5070"}
	fports := []string{"7000", "7100", "7200"}

	running := make([]<-chan *exec.Cmd, 0)
	for i := 0; i <= 2; i++ {
		o := make(options)
		o["-cport"] = cports[i]
		o["-dport"] = dports[i]
		o["-fport"] = fports[i]
		running = append(running, runSnfs(snfs(o)))
	}

	start := time.Now()
	for cmd := range merge(running...) {
		fmt.Printf("pid=%d duration=%s \n", cmd.Process.Pid, time.Since(start))
	}

	//cmd := exec.Command("snfsd")
	//
	//time.AfterFunc(10*time.Second, func() { cmd.Process.Kill() })
	//err := cmd.Run()
	//fmt.Printf("pid=%d duration=%s err=%s\n", cmd.Process.Pid, time.Since(start), err)
}

func snfs(o options) *exec.Cmd {
	log.Println("Starting with ", o)
	return exec.Command("snfsd", o.Args()...)
}

func runSnfs(cmd *exec.Cmd) <-chan *exec.Cmd {
	out := make(chan *exec.Cmd)
	go func() {
		err := cmd.Run()
		if err != nil {
			log.Printf("Process: %d exited.", cmd.Process.Pid)
		}
		out <- cmd
	}()

	return out
}

func merge(cs ...<-chan *exec.Cmd) <-chan *exec.Cmd {
	var wg sync.WaitGroup
	wg.Add(len(cs))

	out := make(chan *exec.Cmd)
	for _, c := range cs {
		go func(c  <-chan *exec.Cmd) {
			exitedCommand := <- c
			out <- exitedCommand
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()


	return out
}
