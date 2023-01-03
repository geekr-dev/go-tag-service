package main

import (
	"log"
	"net"

	pb "github.com/geekr-dev/go-tag-service/proto"
	"github.com/geekr-dev/go-tag-service/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())
	reflection.Register(s)

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("server.serve err: %v", err)
	}
}
