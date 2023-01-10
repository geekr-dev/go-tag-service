package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"

	pb "github.com/geekr-dev/go-tag-service/proto"
	"github.com/geekr-dev/go-tag-service/server"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var port string

type httpError struct {
	Code    int32  `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func grpcGatewayError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}
	httpError := httpError{Code: int32(s.Code()), Message: s.Message()}
	details := s.Details()
	for _, detail := range details {
		if v, ok := detail.(*pb.Error); ok {
			httpError.Code = v.Code
			httpError.Message = v.Message
		}
	}
	resp, _ := json.Marshal(httpError)
	w.Header().Set("Content-Type", marshaler.ContentType(resp))
	w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	_, _ = w.Write(resp)
}

func init() {
	flag.StringVar(&port, "port", "9000", "启动端口号")
	flag.Parse()
}

func main() {
	err := startGrpcGateway(port)
	if err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}

// 分流 grpc 和 http 请求
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

// gRPC 服务器
func initGrpcServer() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())
	reflection.Register(s)
	return s
}

// HTTP 服务器
func initHttpServeMux() *http.ServeMux {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	return serveMux
}

// grpc-gateway 网关
func startGrpcGateway(port string) error {
	httpMux := initHttpServeMux()
	endpoint := "0.0.0.0:" + port
	gwmux := runtime.NewServeMux(runtime.WithErrorHandler(grpcGatewayError))
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	_ = pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gwmux, endpoint, dopts)
	httpMux.Handle("/", gwmux)

	grpcServer := initGrpcServer()
	return http.ListenAndServe(":"+port, grpcHandlerFunc(grpcServer, httpMux))
}
