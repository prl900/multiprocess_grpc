package main

import (
	"fmt"
	"time"
)

type Request struct {
	Value int
	Resp  chan int
}

func Double(reqs chan Request) {
	for req := range reqs {
		time.Sleep(time.Second * 2)
		req.Resp <- 2 * req.Value
	}
}

func main() {
	inChan := make(chan Request)
	go Double(inChan)
	go Double(inChan)

	req1 := Request{4, make(chan int)}
	go func() {
		inChan <- req1
	}()
	req2 := Request{32, make(chan int)}
	go func() {
		inChan <- req2
	}()
	select {
	case out := <-req1.Resp:
		fmt.Println(out)
	}

	select {
	case out := <-req2.Resp:
		fmt.Println(out)
	}
}
