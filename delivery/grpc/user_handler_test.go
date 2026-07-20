package delivery_grpc

import (
	"context"
	"errors"
	"testing"
	"user-service/domain"
	"user-service/mocks"
	userpb "user-service/proto/v1"
	"user-service/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
type TestUserHandler struct {
	name string
	wantErr bool
	wantCode codes.Code
	req *userpb.RegisterRequest
	isTaken string
	EmailEmpty bool
	UsernameEmpty bool
	PasswordLessThen8 bool
	isInternalServerError bool
}

type UserHandlerTest struct {
	suite.Suite
	uc *mocks.MockUserUsecase
	ctrl *gomock.Controller
}
func (s *UserHandlerTest) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	mockUsecse := mocks.NewMockUserUsecase(s.ctrl)
	s.uc = mockUsecse
}
func (s *UserHandlerTest) TestUserHandler() {
	t := s.T()
	ctx := context.Background()
	tests := []TestUserHandler{
		{
			name: "sucess - register",
			wantErr: false,
			req: &userpb.RegisterRequest{
				Username: "joko",
				Email: "joko@example.com",
				Password: "12345678",
			},
			wantCode: codes.OK,
		},
		{
			name: "failed - Email empty",
			wantErr: true,
			req: &userpb.RegisterRequest{
				Username: "joko",
				Email: "",
				Password: "12345678",
			},
			EmailEmpty: true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "failed - username empty",
			wantErr: true,
			req: &userpb.RegisterRequest{
				Username: "",
				Email: "joko@example.com",
				Password: "12345678",
			},
			UsernameEmpty: true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "failed - password len than 8 chars",
			wantErr: true,
			req: &userpb.RegisterRequest{
				Username: "joko",
				Email: "joko@example.com",
				Password: "1234567",
			},
			PasswordLessThen8: true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "failed - email taken",
			wantErr: true,
			isTaken: "email",
			req: &userpb.RegisterRequest{
				Username: "joko",
				Email: "joko@example.com",
				Password: "12345678",
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "failed - username taken",
			wantErr: true,
			isTaken: "username",
			req: &userpb.RegisterRequest{
				Username: "joko",
				Email: "joko@example.com",
				Password: "12345678",
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "failed - internal server error",
			wantErr: true,
			req: &userpb.RegisterRequest{
				Username: "joko",
				Email: "joko@example.com",
				Password: "12345678",
			},
			isInternalServerError: true,
			wantCode: codes.Internal,
		},
		{
			name: "failed - bcrypt error",
			wantErr: true,
			req: &userpb.RegisterRequest{
				Username: "joko",
				Email: "joko@example.com",
				Password: "12345678",
			},
			wantCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.isTaken == "email" {
					s.uc.EXPECT().
					Register(ctx, "joko", "joko@example.com", "12345678").
					Return(domain.User{}, domain.ErrEmailTaken).Times(1)
				} else if tt.isTaken == "username" {
					s.uc.EXPECT().
					Register(ctx, "joko", "joko@example.com", "12345678").
					Return(domain.User{}, domain.ErrUsernameTaken).Times(1)
				} else if tt.EmailEmpty {
					s.uc.EXPECT().
					Register(ctx, "joko", "", "12345678").
					Return(domain.User{}, usecase.ErrEmailEmpty).Times(1)
				} else if tt.UsernameEmpty {
					s.uc.EXPECT().
					Register(ctx, "", "joko@example.com", "12345678").
					Return(domain.User{}, usecase.ErrUsernameEmpty).Times(1)
				} else if tt.PasswordLessThen8 {
					s.uc.EXPECT().
					Register(ctx, "joko", "joko@example.com", "1234567").
					Return(domain.User{}, usecase.ErrPasswordTooShort).Times(1)
				} else if tt.isInternalServerError {
					s.uc.EXPECT().
					Register(ctx, "joko", "joko@example.com", "12345678").
					Return(domain.User{}, usecase.ErrInternalServerError).Times(1)
				} else {
					s.uc.EXPECT().
					Register(ctx, "joko", "joko@example.com", "12345678").
					Return(domain.User{}, errors.New("error from bcrypt")).Times(1)
				}
			} else {
				s.uc.EXPECT().
				Register(ctx, "joko", "joko@example.com", "12345678").
				Return(domain.User{
					ID: 1,
					Username: "joko",
					Email: "joko@example.com",
				}, nil).Times(1)
			}
			
			handler := NewUserHandler(s.uc)
			resp, err := handler.Register(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, st.Code(), tt.wantCode)
				assert.Nil(t, resp)
				return
			}
			assert.Equal(t, resp.Email, tt.req.Email)
		})
	}
}

func (s *UserHandlerTest) TestLogin() {
	t := s.T()
	ctx := context.Background()
	t.Run("success - login", func(t *testing.T) {
		s.uc.EXPECT().
		Login(ctx, "joko", "12345678").
		Return("1234", "4567", nil).Times(1)
		uh := NewUserHandler(s.uc)
		resp, err := uh.Login(ctx, &userpb.LoginRequest{
			Username: "joko",
			Password: "12345678",
		})
		assert.NoError(t, err)
		assert.Equal(t, resp.AccessToken, "1234")
		assert.Equal(t, resp.RefreshToken, "4567")
	})
	t.Run("failed - username is empty", func(t *testing.T) {
		uh := NewUserHandler(s.uc)
		resp, err := uh.Login(ctx, &userpb.LoginRequest{
			Username: "",
			Password: "12345678",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.InvalidArgument)
		assert.Equal(t, st.Message(), "username is required")
	})
	t.Run("failed - password is empty", func(t *testing.T) {
		uh := NewUserHandler(s.uc)
		resp, err := uh.Login(ctx, &userpb.LoginRequest{
			Username: "joko",
			Password: "",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.InvalidArgument)
		assert.Equal(t, st.Message(), "password is required")
	})
	t.Run("failed - password is empty", func(t *testing.T) {
		s.uc.EXPECT().
		Login(ctx, "joko", "12345678").
		Return("", "", usecase.ErrInvalidCredentials).Times(1)
		uh := NewUserHandler(s.uc)
		resp, err := uh.Login(ctx, &userpb.LoginRequest{
			Username: "joko",
			Password: "12345678",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Equal(t, st.Message(), "invalid credential")
	})
	t.Run("failed - unexpected error", func(t *testing.T) {
		s.uc.EXPECT().
		Login(ctx, "joko", "12345678").
		Return("", "", errors.New("unexpected error")).Times(1)
		uh := NewUserHandler(s.uc)
		resp, err := uh.Login(ctx, &userpb.LoginRequest{
			Username: "joko",
			Password: "12345678",
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.Internal)
		assert.Equal(t, st.Message(), "internal server error")
	})
}
func (s *UserHandlerTest) TestRefreshToken() {
	t := s.T()
	ctx := context.Background()
	t.Run("success - get access token", func(t *testing.T) {
		s.uc.EXPECT().
			RefreshAccessToken(gomock.Any(), gomock.Any()).
			Return("ini-access-token-baru", nil).
			Times(1)
		uh := NewUserHandler(s.uc)
		resp, err := uh.RefreshAccessToken(ctx, &userpb.RefreshAccessTokenRequest{
			RefreshToken: "12345",
		})
		assert.NoError(t, err)
		assert.Equal(t, "ini-access-token-baru", resp.AccessToken )
	})
	t.Run("failed - refresh token is empty", func(t *testing.T) {
		uh := NewUserHandler(s.uc)
		resp, err := uh.RefreshAccessToken(ctx, &userpb.RefreshAccessTokenRequest{
			RefreshToken: "",
		})
		assert.Error(t, err)
		assert.Empty(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.InvalidArgument)
	})
	t.Run("failed - expired token", func(t *testing.T) {
		s.uc.EXPECT().
			RefreshAccessToken(gomock.Any(), gomock.Any()).
			Return("", usecase.ErrTokenExpired).
			Times(1)
		uh := NewUserHandler(s.uc)
		resp, err := uh.RefreshAccessToken(ctx, &userpb.RefreshAccessTokenRequest{
			RefreshToken: "12345",
		})
		assert.Error(t, err)
		assert.Empty(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Equal(t, st.Message(), "expired token")
	})
	t.Run("failed - invalid token", func(t *testing.T) {
		s.uc.EXPECT().
			RefreshAccessToken(gomock.Any(), gomock.Any()).
			Return("", usecase.ErrInvalidToken).
			Times(1)
		uh := NewUserHandler(s.uc)
		resp, err := uh.RefreshAccessToken(ctx, &userpb.RefreshAccessTokenRequest{
			RefreshToken: "12345",
		})
		assert.Error(t, err)
		assert.Empty(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.Unauthenticated)
		assert.Equal(t, st.Message(), "invalid token")
	})
	t.Run("failed - unexpected codes", func(t *testing.T) {
		s.uc.EXPECT().
			RefreshAccessToken(gomock.Any(), gomock.Any()).
			Return("", errors.New("internal server error")).
			Times(1)
		uh := NewUserHandler(s.uc)
		resp, err := uh.RefreshAccessToken(ctx, &userpb.RefreshAccessTokenRequest{
			RefreshToken: "12345",
		})
		assert.Error(t, err)
		assert.Empty(t, resp)
		st := status.Convert(err)
		assert.Equal(t, st.Code(), codes.Internal)
		assert.Equal(t, st.Message(), "internal server error")
	})
}
func (s *UserHandlerTest) TearDownSuite() {
	s.ctrl.Finish()
}
func TestUserHandlerTest(t *testing.T) {
	suite.Run(t, new(UserHandlerTest))
}