package unixsock

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ProcessPool struct {
	Pool      []*Process
	TaskQueue chan Task
	Error     chan error
	Health    chan *HealthMsg
}

func (p *ProcessPool) AddProcess(errChan chan error, healthChan chan *HealthMsg) {
	proc := NewProcess(context.Background(), p.TaskQueue, "./process_server", errChan, healthChan)
	proc.Start()
	p.Pool = append(p.Pool, proc)
}

func (p *ProcessPool) RemoveProcess(address string) {
	newPool := []*Process{}
	for _, proc := range p.Pool {
		if proc.Address != address {
			newPool = append(newPool, proc)
		}
	}
	p.Pool = newPool
}

func CreateProcessPool(n int) *ProcessPool {
	p := &ProcessPool{[]*Process{}, make(chan Task), make(chan error), make(chan *HealthMsg)}
	for i := 0; i < n; i++ {
		p.AddProcess(p.Error, p.Health)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-signals:
				p.DeleteProcessPool()
				time.Sleep(1 * time.Second)
				os.Exit(1)
			}
		}
	}()

	go func() {
		for {
			select {
			case err := <-p.Error:
				log.Printf("%v", err)
			}
		}
	}()

	go func() {
		for {
			select {
			case hMsg := <-p.Health:
				p.RemoveProcess(hMsg.Address)
				if hMsg.Replace == true {
					p.AddProcess(p.Error, p.Health)
				}
			}
		}
	}()

	return p
}

func (p *ProcessPool) DeleteProcessPool() {
	for _, proc := range p.Pool {
		proc.Cancel()
	}
}
