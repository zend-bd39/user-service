package usecase

import (
	"context"
	"errors"
	"log"
	"user-service/domain"
	"user-service/pkg"

	"golang.org/x/crypto/bcrypt"
)
var (
	ErrUsernameEmpty     = errors.New("username tidak boleh kosong")
	ErrEmailEmpty        = errors.New("email tidak boleh kosong")
	ErrPasswordTooShort  = errors.New("password minimal 8 karakter")
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrTokenExpired = errors.New("Token expired")
	ErrInvalidToken = errors.New("invalid token")
)

type userUsecase struct {
	repo domain.UserRepository
	jwtService *pkg.JWTService
}

var _ domain.UserUsecase = (*userUsecase)(nil)

func NewUserUsecase(repo domain.UserRepository, jwtSrv *pkg.JWTService) *userUsecase {
	return &userUsecase{
		repo: repo,
		jwtService: jwtSrv,
	}
}

func (u *userUsecase) Register(ctx context.Context, username, email, password string) (domain.User, error) {
	if username == "" {
		return domain.User{}, ErrUsernameEmpty
	}
	if email == "" {
		return domain.User{}, ErrEmailEmpty
	}
	if len(password) < 8 {
		return domain.User{}, ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	user := domain.User{Username: username, Email: email, PasswordHash: string(hash)}
	created, err := u.repo.Create(ctx, user)
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

func (u *userUsecase) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	user, err  := u.repo.FindByUsername(ctx, username)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound){
		return "", "", ErrInternalServerError
	}
	hashToCompare := user.PasswordHash
	if errors.Is(err, domain.ErrUserNotFound) {
		hashToCompare = pkg.CreateDummyHash()
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashToCompare), []byte(password))
	if err != nil {
		return "", "", ErrInvalidCredentials
	}
	accessToken, err = u.jwtService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", ErrInternalServerError
	}
	refreshToken, err = u.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", ErrInternalServerError
	}
	return accessToken, refreshToken, nil
}

func (u *userUsecase) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {

	claims, err := u.jwtService.VerifyToken(refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, pkg.ErrExpiredToken):
			return "", ErrTokenExpired
		case errors.Is(err, pkg.ErrInvalidToken):
			return "", ErrInvalidToken
		default:
			return "", ErrInternalServerError
		}
	}
	
	if claims.TokenType != "refresh" {
		log.Printf("[RefreshAccessToken] gagal FindByID untuk user %d: %v", claims.UserID, err)
		return "", ErrInvalidToken
	}
	user, err := u.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", ErrInvalidToken
	}
	return u.jwtService.GenerateAccessToken(user.ID, user.Role)
}