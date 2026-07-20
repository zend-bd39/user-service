package delivery_grpc

import (
	"context"
	"errors"
	"log/slog"
	"user-service/domain"
	userpb "user-service/proto/v1"
	"user-service/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	userUC domain.UserUsecase
}
var _ userpb.UserServiceServer = (*UserHandler)(nil)

func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{
		userUC: uc,
	}
}
func (h *UserHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	user, err  := h.userUC.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		st := status.Convert(err)
		switch {
		case errors.Is(err, usecase.ErrEmailEmpty):
			slog.Error("Error" , slog.Any("codes.InvalidArgument", uint32(st.Code())))
			return nil, status.Error(codes.InvalidArgument, "email cant empty")
		case errors.Is(err, usecase.ErrUsernameEmpty):
			return nil, status.Error(codes.InvalidArgument, "username cant empty")
		case errors.Is(err, usecase.ErrPasswordTooShort):
			return nil, status.Error(codes.InvalidArgument, "password min 8 chars")
		case errors.Is(err, domain.ErrEmailTaken):
			return nil, status.Error(codes.AlreadyExists, "email is taken")
		case errors.Is(err, domain.ErrUsernameTaken):
			return nil, status.Error(codes.AlreadyExists, "username is taken")
		case errors.Is(err, usecase.ErrInternalServerError):
			return nil, status.Error(codes.Internal, "internal server error")
		default:
			slog.Error("Bcrypt error", slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &userpb.RegisterResponse{
		Id: int32(user.ID),
		Username: user.Username,
		Email: user.Email,
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	access,refresh, err  := h.userUC.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credential")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &userpb.LoginResponse{
		AccessToken: access,
		RefreshToken: refresh,
	}, nil
}

func (h *UserHandler) RefreshAccessToken(ctx context.Context, req *userpb.RefreshAccessTokenRequest) ( *userpb.RefreshAccessTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}
	access, err  := h.userUC.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidToken):
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		case errors.Is(err, usecase.ErrTokenExpired):
			return nil, status.Error(codes.Unauthenticated, "expired token")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &userpb.RefreshAccessTokenResponse{
		AccessToken: access,
	}, nil
}