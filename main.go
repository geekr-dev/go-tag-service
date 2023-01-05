package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	pb "github.com/geekr-dev/go-tag-service/proto"
	"github.com/geekr-dev/go-tag-service/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var grpcPort string
var httpPort string

func init() {
	flag.StringVar(&grpcPort, "grpc_port", "9000", "gRPC 启动端口")
	flag.StringVar(&httpPort, "http_port", "8001", "HTTP 启动端口")
	flag.Parse()
}

func main() {

	errs := make(chan error)
	go func() {
		err := startHttpServer(httpPort)
		if err != nil {
			errs <- err
		}
	}()

	go func() {
		err := startGrpcServer(grpcPort)
		if err != nil {
			errs <- err
		}
	}()

	err := <-errs
	if err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}

// 启动 HTTP 服务器
func startHttpServer(port string) error {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	return http.ListenAndServe(":"+port, serverMux)
}

// 启动 gRPC 服务器
func startGrpcServer(port string) error {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())
	reflection.Register(s)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	return s.Serve(lis)
}
