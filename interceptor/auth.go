package interceptor

import (
	"context"
	"user-service/pkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)
type contextKey string
const ClaimsContextKey contextKey = "claims"

func AuthInterceptor(jwtService *pkg.JWTService, publicMethod map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if publicMethod[info.FullMethod] {
			// endpoint publik (Register, Login) skip auth
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not found")
		}
		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "token not found")
		}
		claims, err := jwtService.VerifyToken(tokens[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if claims.TokenType != "access" {
			return nil, status.Error(codes.Unauthenticated, "invalid token type")
		}
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)
		return handler(ctx, req)
	}
}