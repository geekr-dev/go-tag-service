package main

import (
	"context"
	"log"

	"github.com/geekr-dev/go-blog-app/pkg/tracer"
	"github.com/geekr-dev/go-tag-service/global"
	"github.com/geekr-dev/go-tag-service/internal/middleware"
	"github.com/geekr-dev/go-tag-service/pkg/registry/etcd"
	pb "github.com/geekr-dev/go-tag-service/proto"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Auth struct {
	AppKey    string
	AppSecret string
}

func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"app_key": a.AppKey, "app_secret": a.AppSecret}, nil
}

func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func init() {
	// 初始化链路追踪
	err := setupTracer()
	if err != nil {
		log.Fatalf("init.setupTracer err: %v", err)
	}
}

func setupTracer() error {
	jaegerTracer, _, err := tracer.NewJaegerTracer("go-tag-service", "localhost:6831")
	if err != nil {
		return err
	}
	global.Tracer = jaegerTracer
	return nil
}

func main() {
	auth := Auth{
		AppKey:    "geekr-dev",
		AppSecret: "go-tag-service",
	}
	ctx := context.Background()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(&auth),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				// 超时控制
				middleware.UnaryContextTimeout(),
				// 最大尝试次数2次，在指定错误码下进行重试
				grpc_retry.UnaryClientInterceptor(
					grpc_retry.WithMax(2),
					grpc_retry.WithCodes(
						codes.Unknown,
						codes.Internal,
						codes.DeadlineExceeded,
					),
				),
				// 客户端拦截器实现服务追踪
				middleware.ClientTracing(),
			),
		),
	}
	// 服务发现
	etcdRegistry, err := etcd.New()
	if err != nil {
		log.Fatalf("create etcd client failed: %v", err)
	}
	// defer etcdRegistry.Close()
	endpoint, err := etcdRegistry.GetEndpoint(global.SERVICE_NAME)
	if err != nil {
		log.Fatalf("get endpoint from registry failed: %v", err)
	}
	if endpoint == "" {
		log.Fatalln("no available endpoints found")
	}
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer conn.Close()

	tagServiceClient := pb.NewTagServiceClient(conn)
	resp, err := tagServiceClient.GetTagList(ctx, &pb.GetTagListRequest{Name: "Go"})
	if err != nil {
		log.Fatalf("tagServiceClient.GetTagList err: %v", err)
	}
	log.Printf("resp: %v", resp)
}
