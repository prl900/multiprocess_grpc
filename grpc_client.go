package main

import (
	pb "./unixsock"
	"flag"
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("p", 6000, "gRPC server listening port.")
	flag.Parse()

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", *port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewComputeClient(conn)

	r, err := c.Double(context.Background(), &pb.Obj{456})
	if err != nil {
		log.Fatalf("could not double: %v", err)
	}
	log.Printf("Double of 456 is : %d", r.Value)
}
