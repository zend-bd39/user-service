package usecase

import (
	"context"
	"errors"
	"slices"
	"user-service/domain"

	"golang.org/x/crypto/bcrypt"
)

type userUsecaseMid struct {
	userRepo domain.UserRepository
	registerExecutor RegisterFn
}

type RegisterFn func(ctx context.Context, username, email, password string) (domain.User, error)
type RegisterMiddleware func(RegisterFn) RegisterFn

func (u *userUsecaseMid) Register(ctx context.Context, username, email, password string) (domain.User, error) {
	return u.registerExecutor(ctx, username, email, password)
}
func NewUserUsecaseMid(repo domain.UserRepository, regMids ...RegisterMiddleware) *userUsecaseMid {
	regCore := func(ctx context.Context, username, email, password string) (domain.User, error) {
		user := domain.User{
			Username: username,
			Email: email,
			PasswordHash: password,
		}
		created, err := repo.Create(ctx, user)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrEmailTaken):
				return domain.User{}, domain.ErrEmailTaken
			case errors.Is(err, domain.ErrUsernameTaken):
				return domain.User{}, domain.ErrUsernameTaken
			default:
				return domain.User{}, ErrInternalServerError
			}
		}
		return created, nil
	}
	for _, m := range slices.Backward(regMids) {
		regCore = m(regCore)
	}
	
	return &userUsecaseMid{
		userRepo: repo,
		registerExecutor: regCore,
	}
}
func HashPassword(next RegisterFn) RegisterFn {
	return func(ctx context.Context, username, email, password string) (domain.User, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return domain.User{}, err
		}
		return next(ctx, username, email, string(hash))
	}
}

func ValidationInput(next RegisterFn) RegisterFn {
	return func(ctx context.Context, username, email, password string) (domain.User, error) {
		if username == "" {
			return domain.User{}, ErrUsernameEmpty
		}
		if email == "" {
			return domain.User{}, ErrEmailEmpty
		}
		if len(password) < 8 {
			return domain.User{}, ErrPasswordTooShort
		}
		return next(ctx, username, email, password)
	}
}