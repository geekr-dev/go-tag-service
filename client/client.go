package main

import (
	"context"
	"log"

	"github.com/geekr-dev/go-tag-service/internal/middleware"
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
			),
		),
	}
	conn, err := grpc.DialContext(ctx, "localhost:9000", opts...)
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
