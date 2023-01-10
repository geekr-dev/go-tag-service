package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/geekr-dev/go-tag-service/pkg/bapi"
	"github.com/geekr-dev/go-tag-service/pkg/errcode"
	pb "github.com/geekr-dev/go-tag-service/proto"
)

type TagServer struct {
	*pb.UnimplementedTagServiceServer
}

func NewTagServer() *TagServer {
	return &TagServer{}
}

func (ts *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	api := bapi.NewAPI("http://127.0.0.1:8000")
	body, err := api.GetTagList(ctx, r.GetName())
	if err != nil {
		return nil, errcode.TogRPCError(errcode.ERROR_GET_TAG_LIST_FAIL)
	}
	time.Sleep(60 * time.Second) // 模拟耗时60s
	tagList := pb.GetTagListReply{}
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, errcode.TogRPCError(errcode.Fail)
	}

	return &tagList, nil
}
