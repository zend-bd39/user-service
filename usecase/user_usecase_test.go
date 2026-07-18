package usecase

import (
	"context"
	"testing"
	"time"
	"user-service/domain"
	"user-service/mocks"
	"user-service/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecaseTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	mockRepo *mocks.MockUserRepository
	usecase *userUsecase
	jwtService *pkg.JWTService
}

func (suite *UserUsecaseTestSuite) SetupSuite() {

	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = mocks.NewMockUserRepository(suite.ctrl)
	suite.jwtService = pkg.NewJWTService("test-123", time.Second * 1, time.Second * 2)
	suite.usecase = NewUserUsecase(suite.mockRepo, suite.jwtService)
}

func (suite *UserUsecaseTestSuite) TestRegister() {
	t := suite.T()
	ctx := context.Background()
	t.Run("success - register", func(t *testing.T) {
		suite.mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func (ctx context.Context, user domain.User) (domain.User, error)  {
				err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
				assert.NoError(t, err)
				user.ID = 1
				return user, nil
			}).Times(1)
		created, err := suite.usecase.Register(ctx, "budi", "budi@example.com", "password123")
		assert.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, "budi", created.Username)
	})
	t.Run("failed - username is empty", func(t *testing.T) {
		suite.mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).Times(0)
		created, err := suite.usecase.Register(ctx, "", "budi@example.com", "password123")
		assert.Error(t, err)
		assert.Empty(t, created)
	})
	t.Run("failed - email is empty", func(t *testing.T) {
		suite.mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).Times(0)
		created, err := suite.usecase.Register(ctx, "budi", "", "password123")
		assert.Error(t, err)
		assert.Empty(t, created)
	})
	t.Run("failed - password is short", func(t *testing.T) {
		suite.mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).Times(0)
		created, err := suite.usecase.Register(ctx, "budi", "budi@example.com", "pas")
		assert.Error(t, err)
		assert.Empty(t, created)
	})
	t.Run("failed - email is taken", func(t *testing.T) {
		suite.mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(domain.User{}, domain.ErrEmailTaken).
			Times(1)
		created, err := suite.usecase.Register(ctx, "budi", "budi@example.com", "password123")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrEmailTaken, err)
		assert.Empty(t, created)
	})
	t.Run("failed - username is taken", func(t *testing.T) {
		suite.mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(domain.User{}, domain.ErrUsernameTaken).
			Times(1)
		created, err := suite.usecase.Register(ctx, "budi", "budi@example.com", "password123")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUsernameTaken, err)
		assert.Empty(t, created)
	})
	t.Run("failed - internal", func(t *testing.T) {
		suite.mockRepo.
			EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(domain.User{}, ErrInternalServerError).
			Times(1)
		created, err := suite.usecase.Register(ctx, "budi", "budi@example.com", "password123")
		assert.Error(t, err)
		assert.ErrorIs(t, ErrInternalServerError, err)
		assert.Empty(t, created)
	})
}

func (suite *UserUsecaseTestSuite) TestLogin() {
	t:= suite.T()
	ctx := context.Background()
	t.Run("success - login", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("joko123"), bcrypt.DefaultCost)
		suite.mockRepo.EXPECT().
			FindByUsername(ctx, "joko").
			Return(domain.User{
				ID: 1,
				Username: "joko",
				Email: "joko@example.com",
				PasswordHash: string(hash),
				Role: "user",
			}, nil)
		access, refresh, err  := suite.usecase.Login(ctx, "joko", "joko123")
		assert.NoError(t, err)
		assert.NotEmpty(t, access)
		assert.NotEmpty(t, refresh)
	})
	t.Run("failed - Not found", func(t *testing.T) {
		suite.mockRepo.EXPECT().
			FindByUsername(ctx, "joko").
			Return(domain.User{}, domain.ErrUserNotFound)
		access, refresh, err  := suite.usecase.Login(ctx, "joko", "joko123")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})
	t.Run("failed - unexpected error", func(t *testing.T) {
		suite.mockRepo.EXPECT().
			FindByUsername(ctx, "joko").
			Return(domain.User{}, ErrInternalServerError)
		access, refresh, err  := suite.usecase.Login(ctx, "joko", "joko123")
		assert.Error(t, err)
		assert.ErrorIs(t, ErrInternalServerError, err)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})
	t.Run("failed - Error from bcrypt", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("joko123"), bcrypt.DefaultCost)
		suite.mockRepo.EXPECT().
			FindByUsername(ctx, "joko").
			Return(domain.User{
				ID: 1,
				Username: "joko",
				Email: "joko@example.com",
				PasswordHash: string(hash),
				Role: "user",
			}, nil)
		access, refresh, err  := suite.usecase.Login(ctx, "joko", "joko1234")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})
}
func (suite *UserUsecaseTestSuite) TestRefreshAccessToken() {
	t:= suite.T()
	ctx := context.Background()
	t.Run("Success - Generate new access token", func(t *testing.T) {
		suite.mockRepo.EXPECT().
			FindByID(ctx,1).
			Return(domain.User{
				ID: 1,
				Role: "user",
			}, nil).Times(1)
		refresh, err := suite.jwtService.GenerateRefreshToken(1)
		assert.NoError(t, err)
		access, err := suite.usecase.RefreshAccessToken(ctx, refresh)
		assert.NoError(t, err)
		assert.NotEmpty(t, access)
	})

	t.Run("Failed - Error verify token - Invalid", func(t *testing.T) {
		refresh := "123"
		access, err := suite.usecase.RefreshAccessToken(ctx, refresh)
		assert.Error(t, err)
		assert.Empty(t, access)
		assert.Equal(t, ErrInvalidToken, err)
	})
	t.Run("Failed - Error verify token - Expired", func(t *testing.T) {
		refresh, err := suite.jwtService.GenerateRefreshToken(1)
		time.Sleep(time.Second * 3)
		access, err := suite.usecase.RefreshAccessToken(ctx, refresh)
		assert.Error(t, err)
		assert.Empty(t, access)
		assert.Equal(t, ErrTokenExpired, err)
	})
	t.Run("Failed - Error Invalid Token Type", func(t *testing.T) {
		access, err := suite.jwtService.GenerateAccessToken(1, "user")
		newAccess, err := suite.usecase.RefreshAccessToken(ctx, access)
		assert.Error(t, err)
		assert.Empty(t, newAccess)
		assert.Equal(t, ErrInvalidToken, err)
	})
}

func (suite *UserUsecaseTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}
func TestUserUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseTestSuite))
}