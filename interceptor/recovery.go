package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			if r := recover(); r != nil {
				duration := time.Since(start)
				log.Printf("[gRPC SERVER PANIC] %s | recover: %v | duration: %s", info.FullMethod, r, duration)
				err = status.Error(codes.Internal, "unexpected error happened")
				return
			}
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}