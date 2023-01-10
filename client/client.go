package main

import (
	"context"
	"log"

	"github.com/geekr-dev/go-tag-service/internal/middleware"
	pb "github.com/geekr-dev/go-tag-service/proto"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				middleware.UnaryContextTimeout(),
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
