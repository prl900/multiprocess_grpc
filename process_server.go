package main

import (
	"encoding/gob"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	us "./unixsock"
)

func dataHandler(enc *gob.Encoder, dec *gob.Decoder) {
	// Create a decoder and receive a value.
	var o us.Obj
	err := dec.Decode(&o)
	if err != nil {
		log.Fatal("decode:", err)
	}

	time.Sleep(2 * time.Second)

	err = enc.Encode(us.Obj{2 * o.Value})
	if err != nil {
		log.Fatal("encode:", err)
	}
}

func main() {
	conn := os.Args[1]
	l, err := net.Listen("unix", conn)
	if err != nil {
		log.Fatal(err)
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-signals:
				log.Println("Caught SIGHUP, reloading config...")
			}
		}
	}()

	for {
		fd, err := l.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		dec := gob.NewDecoder(fd)
		enc := gob.NewEncoder(fd)

		dataHandler(enc, dec)
		fd.Close()
	}
}
