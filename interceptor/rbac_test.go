package interceptor

import (
	"context"
	"testing"
	"user-service/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RBACInterceptorTestSuite struct {
	suite.Suite
	publicRoles map[string][]string
}

func (s *RBACInterceptorTestSuite) SetupSuite() {
	s.publicRoles = map[string][]string{
		"/proto.v1.UserService/Admin": {"admin"},
	}
}

func (s *RBACInterceptorTestSuite) TestRBAC() {
	t := s.T()
	ctx := context.Background()
	t.Run("success - public method", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/Register",
		}
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			handlerCalled = true
			return "mock-response", nil

		}
		interCp := RBACInterceptor(s.publicRoles)
		resp, err := interCp(ctx, nil, info, mockHandler)
		assert.NoError(t, err)
		assert.True(t, handlerCalled, "handler utama harus terpanggil untuk public method register")
		assert.Equal(t, "mock-response", resp)
	})
	t.Run("failed - not claims", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/Admin",
		}
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}
		interCp := RBACInterceptor(s.publicRoles)
		resp, err := interCp(ctx, nil, info, mockHandler)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Equal(t, st.Message(), "claims not found")
		assert.False(t, handlerCalled, "handler tidak dipanggil")
		assert.Nil(t, resp)
	})
	t.Run("success - role is admin", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/Admin",
		}
		ctxValue := context.WithValue(ctx, ClaimsContextKey, &pkg.Claims{
			UserID: 1,
			Role: "admin",
		})
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			handlerCalled = true
			return "mock response", nil
		}
		interCp := RBACInterceptor(s.publicRoles)
		resp, err := interCp(ctxValue, nil, info, mockHandler)
		assert.NoError(t, err)
		assert.Equal(t, "mock response", resp)
		assert.True(t, handlerCalled)
	})

	t.Run("failed - role is not admin", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/Admin",
		}
		ctxValue := context.WithValue(ctx, ClaimsContextKey, &pkg.Claims{
			UserID: 1,
			Role: "user",
		})
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			handlerCalled = true
			return "mock response", nil
		}
		interCp := RBACInterceptor(s.publicRoles)
		resp, err := interCp(ctxValue, nil, info, mockHandler)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, st.Code(), codes.PermissionDenied)
		assert.Equal(t, st.Message(), "access denied for this role")
		assert.False(t, handlerCalled, "handler tidak dipanggil")
		assert.Nil(t, resp)
	})
}
func TestRBACInterceptorTestSuite(t *testing.T) {
	suite.Run(t, new(RBACInterceptorTestSuite))
}