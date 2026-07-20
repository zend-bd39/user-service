package interceptor

import (
	"context"
	"testing"
	"time"
	"user-service/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptorSuiteTest struct {
	suite.Suite
	jwtService *pkg.JWTService
	publicMethod map[string]bool
	publicRoles map[string][]string
}

func (s *AuthInterceptorSuiteTest) SetupSuite() {
	s.jwtService = pkg.NewJWTService(
		"test123",
		1 * time.Second,
		3 * time.Second,
	)
	s.publicMethod = map[string]bool{
		"/proto.v1.UserService/Register": true,
		"/proto.v1.UserService/Login": true,
		"/proto.v1.UserService/RefreshAccessToken":true,
	}
}
func (s *AuthInterceptorSuiteTest) TestAuthInterceptor() {
	t := s.T()
	ctx := context.Background()
	t.Run("success - public method", func(t *testing.T) {
		access, err := s.jwtService.GenerateAccessToken(1, "user")
		assert.NoError(t, err)
		md := metadata.Pairs(
			"authorization", access,
		)
		ctxMD := metadata.NewIncomingContext(ctx, md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/Register",
		}
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			handlerCalled = true
			return "mock-response", nil

		}
		interCp := AuthInterceptor(s.jwtService, s.publicMethod)
		resp, err := interCp(ctxMD, nil, info, mockHandler)
		assert.NoError(t, err)
		assert.True(t, handlerCalled, "handler utama harus terpanggil untuk public method register")
		assert.Equal(t, "mock-response", resp)
	})
	t.Run("failed - metadata not found", func(t *testing.T) {
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/xxx",
		}
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}
		interCp := AuthInterceptor(s.jwtService, s.publicMethod)
		resp, err := interCp(ctx, nil, info, mockHandler)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Nil(t, resp)
		assert.False(t, handlerCalled, "handler tidak boleh dipanggil")
	})
	t.Run("failed - authorization", func(t *testing.T) {
		md := metadata.Pairs(
			"authorization", "token_palsu",
		)
		ctxMD := metadata.NewIncomingContext(ctx, md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/xxx",
		}
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			return nil, nil

		}
		interCp := AuthInterceptor(s.jwtService, s.publicMethod)
		resp, err := interCp(ctxMD, nil, info, mockHandler)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, st.Message(), "invalid token")
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Nil(t, resp)
		assert.False(t, handlerCalled, "handler tidak boleh dipanggil")
	})
	t.Run("failed - token not found", func(t *testing.T) {
		md := metadata.Pairs(
			"bukan_auth", "token_palsu",
		)
		ctxMD := metadata.NewIncomingContext(ctx, md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/xxx",
		}
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}
		interCp := AuthInterceptor(s.jwtService, s.publicMethod)
		resp, err := interCp(ctxMD, nil, info, mockHandler)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, st.Message(), "token not found")
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Nil(t, resp)
		assert.False(t, handlerCalled, "handler tidak boleh dipanggil")
	})
	t.Run("failed - req using token refresh", func(t *testing.T) {
		refresh, err := s.jwtService.GenerateRefreshToken(1)
		assert.NoError(t, err)
		md := metadata.Pairs(
			"authorization", refresh,
		)
		ctxMD := metadata.NewIncomingContext(ctx, md)
		info := &grpc.UnaryServerInfo{
			FullMethod: "/proto.v1.UserService/xxx",
		}
		handlerCalled := false
		mockHandler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}
		interCp := AuthInterceptor(s.jwtService, s.publicMethod)
		resp, err := interCp(ctxMD, nil, info, mockHandler)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, st.Message(), "invalid token type")
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Nil(t, resp)
		assert.False(t, handlerCalled, "handler tidak boleh dipanggil")
	})
}

func TestAuthInterceptorSuiteTest(t *testing.T) {
	suite.Run(t, new(AuthInterceptorSuiteTest))
}