package unixsock

import (
	"context"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

type HealthMsg struct {
	Address string
	Replace bool
}

func createUniqueFilename(dir string) (string, error) {
	var filename string
	var err error

	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return filename, fmt.Errorf("Could not get the current working directory %v", err)
		}
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return filename, fmt.Errorf("The provided directory does not exist %v", err)
	}

	for {
		filename = filepath.Join(dir, strconv.FormatUint(rand.Uint64(), 16))
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return filename, nil
		}
	}
}

type Task struct {
	Payload *Obj
	Resp    chan *Obj
}

type Process struct {
	Context    context.Context
	CancelFunc context.CancelFunc
	TaskQueue  chan Task
	Cmd        *exec.Cmd
	Address    string
	Error      chan error
	Health     chan *HealthMsg
}

func NewProcess(ctx context.Context, tQueue chan Task, binary string, errChan chan error, healthChan chan *HealthMsg) *Process {
	newCtx, cancel := context.WithCancel(ctx)
	address, err := createUniqueFilename("/tmp")
	if err != nil {
		panic(err)
	}

	return &Process{newCtx, cancel, tQueue, exec.CommandContext(newCtx, binary, address), address, errChan, healthChan}
}

func (p *Process) waitReady() error {
	timer := time.NewTimer(time.Millisecond * 200)
	ready := make(chan struct{})

	go func(signal chan struct{}) {
		for {
			time.Sleep(time.Millisecond * 20)
			if _, err := os.Stat(p.Address); err == nil {
				signal <- struct{}{}
				return
			}
		}
	}(ready)

	select {
	case <-ready:
		return nil
	case <-timer.C:
		return fmt.Errorf("Address file creation timed out")
	}
}

func (p *Process) Start() {
	err := p.Cmd.Start()
	if err != nil {
		p.Error <- err
	}

	//log.Printf("Process running with PID %d\n", p.Cmd.Process.Pid)

	err = p.waitReady()
	if err != nil {
		p.Error <- err
	}

	go func() {
		defer p.Cancel()
		var out Obj

		for task := range p.TaskQueue {
			conn, err := net.Dial("unix", p.Address)
			if err != nil {
				p.Error <- fmt.Errorf("dial failed: ", err)
			}

			enc := gob.NewEncoder(conn)
			dec := gob.NewDecoder(conn)

			err = enc.Encode(task.Payload)
			if err != nil {
				p.Error <- fmt.Errorf("encode failed: ", err)
			}

			err = dec.Decode(&out)
			if err != nil {
				p.Error <- fmt.Errorf("decode failed: ", err)
			}

			task.Resp <- &out
		}
	}()

	go func() {
		err := p.Cmd.Wait()
		if err != nil {
			p.Error <- fmt.Errorf("Process %s: %v", p.Address, err)
		}

		select {
		case <-p.Context.Done():
			err := os.Remove(p.Address)
			if err != nil {
				p.Error <- err
			}
			p.Error <- p.Context.Err()
			p.Health <- &HealthMsg{p.Address, false}
		default:
			err = os.Remove(p.Address)
			if err != nil {
				p.Error <- err
			}
			p.Health <- &HealthMsg{p.Address, true}
		}
	}()
}

func (p *Process) Cancel() {
	p.CancelFunc()
	time.Sleep(100 * time.Millisecond)
}
