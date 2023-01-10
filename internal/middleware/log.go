package middleware

import (
	"context"
	"log"
	"time"

	"github.com/geekr-dev/go-tag-service/pkg/errcode"
	"google.golang.org/grpc"
)

// 访问日志
func AccessLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	reqLog := "access request log: method: %s, begin_time: %d, request: %v"
	beginTime := time.Now().Local().Unix()
	log.Printf(reqLog, info.FullMethod, beginTime, req)

	resp, err := handler(ctx, req)

	respLog := "access response log: method: %s, end_time: %d, response: %v"
	endTime := time.Now().Local().Unix()
	log.Printf(respLog, info.FullMethod, endTime, resp)
	return resp, err
}

// 错误日志
func ErrorLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		errLog := "err log: method: %s, code: %v, message: %v, details: %v"
		s := errcode.FromError(err)
		log.Printf(errLog, info.FullMethod, s.Code(), s.Err().Error(), s.Details())
	}
	return resp, err
}
