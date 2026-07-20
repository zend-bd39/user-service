package domain

import (
	"context"
	"errors"
)
var (
	ErrEmailTaken = errors.New("email already exists in database")
	ErrUsernameTaken = errors.New("username already exists in database")
	ErrUserNotFound     = errors.New("user not found")
)
type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Role         string
}


type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	FindByUsername(ctx context.Context, username string) (User, error)
	FindByID(ctx context.Context, id int)(User, error)
}

type UserUsecase interface {
	Register(ctx context.Context, username, email, password string) (User, error)
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
}