package main

import (
	pb "./unixsock"
	"flag"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct {
	Pool *pb.ProcessPool
}

func (s *server) Double(ctx context.Context, in *pb.Obj) (*pb.Obj, error) {
	r := make(chan *pb.Obj)
	s.Pool.TaskQueue <- pb.Task{in, r}

	return <-r, nil
}

func main() {
	port := flag.Int("p", 6000, "gRPC server listening port.")
	poolSize := flag.Int("n", 8, "Maximum number of requests handled concurrently.")
	//aws := flag.Bool("aws", true, "Needs access to AWS S3 rasters?")
	flag.Parse()

	p := pb.CreateProcessPool(*poolSize)

	s := grpc.NewServer()
	pb.RegisterComputeServer(s, &server{Pool: p})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//lis = netutil.LimitListener(lis, *conc)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
