package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func defaultContextTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	// 未设置deadline时将其设置为60s
	if _, ok := ctx.Deadline(); !ok {
		defaultTimeout := 30 * time.Second
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	}
	return ctx, cancel
}

// 一元调用客户端拦截器
func UnaryContextTimeout() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := defaultContextTimeout(ctx)
		if cancel != nil {
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// 流式调用客户端拦截器
func StreamContextTimeout() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, cancel := defaultContextTimeout(ctx)
		if cancel != nil {
			defer cancel()
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
