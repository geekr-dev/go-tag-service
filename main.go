package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	pb "github.com/geekr-dev/go-tag-service/proto"
	"github.com/geekr-dev/go-tag-service/server"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var port string

func init() {
	flag.StringVar(&port, "port", "9000", "启动端口号")
	flag.Parse()
}

func main() {
	// 先监听 TCP 端口，gRPC 和 HTTP 都是基于 TCP 的
	lis, err := startTcpServer(port)
	if err != nil {
		log.Fatalf("start tcp server failed: %v", err)
	}

	m := cmux.New(lis)
	// gRPC 基于 application/grpc 请求头进行分流
	grpcLis := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpLis := m.Match(cmux.HTTP1Fast())

	grpcServer := initGrpcServer()
	httpServer := initHttpServer(port)
	go grpcServer.Serve(grpcLis)
	go httpServer.Serve(httpLis)

	err = m.Serve()
	if err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}

// TCP 服务器
func startTcpServer(port string) (net.Listener, error) {
	return net.Listen("tcp", ":"+port)
}

// gRPC 服务器
func initGrpcServer() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())
	reflection.Register(s)

	return s
}

// HTTP 服务器
func initHttpServer(port string) *http.Server {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	return &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}
}
