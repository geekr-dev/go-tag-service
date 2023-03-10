package server

import (
	"context"
	"encoding/json"

	"github.com/geekr-dev/go-tag-service/pkg/bapi"
	"github.com/geekr-dev/go-tag-service/pkg/errcode"
	pb "github.com/geekr-dev/go-tag-service/proto"

	"google.golang.org/grpc/metadata"
)

type TagServer struct {
	*pb.UnimplementedTagServiceServer
	auth *Auth
}

type Auth struct{}

func (a *Auth) GetAppKey() string {
	return "geekr-dev"
}

func (a *Auth) GetAppSecret() string {
	return "go-tag-service"
}

func (a *Auth) Check(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)

	var appKey, appSecret string
	if value, ok := md["app_key"]; ok {
		appKey = value[0]
	}
	if value, ok := md["app_secret"]; ok {
		appSecret = value[0]
	}
	if appKey != a.GetAppKey() || appSecret != a.GetAppSecret() {
		return errcode.TogRPCError(errcode.Unauthorized)
	}

	return nil
}

func NewTagServer() *TagServer {
	return &TagServer{}
}

func (ts *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	if err := ts.auth.Check(ctx); err != nil {
		return nil, err
	}

	api := bapi.NewAPI("http://127.0.0.1:8000")
	body, err := api.GetTagList(ctx, r.GetName())
	if err != nil {
		return nil, errcode.TogRPCError(errcode.ERROR_GET_TAG_LIST_FAIL)
	}
	// time.Sleep(60 * time.Second) // 模拟耗时60s
	tagList := pb.GetTagListReply{}
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, errcode.TogRPCError(errcode.Fail)
	}

	return &tagList, nil
}
