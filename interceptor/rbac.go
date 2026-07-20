package interceptor

import (
	"context"
	"slices"
	"user-service/pkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RBACInterceptor(requireRole map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		roles, ok := requireRole[info.FullMethod]
		if !ok {
			return handler(ctx, req)
		}
		claims, ok := ctx.Value(ClaimsContextKey).(*pkg.Claims)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "claims not found")
		}
		if slices.Contains(roles, claims.Role) {
			return handler(ctx, req)
		}
		return nil, status.Error(codes.PermissionDenied, "access denied for this role")
	}
}