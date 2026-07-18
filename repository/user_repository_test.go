package repository

import (
	"context"
	"log"
	"testing"
	"user-service/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	pgContainer *PostgresContainer
	repository domain.UserRepository
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	pgContainer, err := CreatePostgresContainer(ctx)
	suite.Require().NoError(err, "failed create postgres container")
	suite.pgContainer = pgContainer
	// create database mandiri
	schemaSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL DEFAULT 'user',
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`
	_, err = suite.pgContainer.Pool.Exec(ctx, schemaSQL)
	suite.Require().NoError(err, "failed to migrate database schema")
	userRepository := NewUserRepository(suite.pgContainer.Pool)
	suite.repository = userRepository
	// _, err = suite.pgContainer.Pool.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
 	// suite.NoError(err)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	ctx := context.Background()
	if suite.pgContainer.Pool != nil {
		suite.pgContainer.Pool.Close()
	}
	if err := suite.pgContainer.Terminate(ctx); err != nil {
		log.Fatalf("error terminating postgres container: %v", err)
	}
}
func (suite *UserRepositoryTestSuite) TestCreateUser() {
	t := suite.T()
	ctx := context.Background()
	t.Run("success create user", func(t *testing.T) {
		user, err := suite.repository.Create(ctx, domain.User{
			Username: "andy",
			Email: "andy@example.com",
			PasswordHash: "ini_hash_password_andy",
		})
		assert.NoError(t, err)
		assert.Equal(t, "andy", user.Username)
		assert.Equal(t, "andy@example.com", user.Email)
	})
	t.Run("failed - Email is taken", func(t *testing.T) {
		user, err := suite.repository.Create(ctx, domain.User{
			Username: "andy123",
			Email: "andy@example.com",
			PasswordHash: "ini_hash_password_andy",
		})
		assert.Error(t, err)
		assert.Empty(t, user)
		assert.Equal(t, domain.ErrEmailTaken, err)
	})
	t.Run("failed - username is taken", func(t *testing.T) {
		user, err := suite.repository.Create(ctx, domain.User{
			Username: "andy",
			Email: "andy11@example.com",
			PasswordHash: "ini_hash_password_andy",
		})
		assert.Error(t, err)
		assert.Empty(t, user)
		assert.Equal(t, domain.ErrUsernameTaken, err)
	})
}
func (suite *UserRepositoryTestSuite) TestFindByID() {
	t := suite.T()
	ctx := context.Background()

	created, err := suite.repository.Create(ctx, domain.User{
		Username: "budi_findbyid",
		Email: "budi_findbyid@example.com",
		PasswordHash: "hash123",
	})
	suite.Require().NoError(err)
	t.Run("success Find ID", func(t *testing.T) {
		user, err := suite.repository.FindByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "budi_findbyid", user.Username)
		assert.Equal(t, "budi_findbyid@example.com", user.Email)
	})
	t.Run("failed - user not found", func(t *testing.T) {
		user, err := suite.repository.FindByID(ctx, 99999)
		assert.Error(t, err)
		assert.Empty(t, user)
		assert.Equal(t, ErrNotFound, err)
	})

}
func (suite *UserRepositoryTestSuite) TestFindByUserName() {
	t := suite.T()
	ctx := context.Background()

	created, err := suite.repository.Create(ctx, domain.User{
		Username: "budi_findbyusername",
		Email: "budi_findbyusername@example.com",
		PasswordHash: "hash123",
	})
	suite.Require().NoError(err)
	t.Run("success Find Username", func(t *testing.T) {
		user, err := suite.repository.FindByUsername(ctx, created.Username)
		assert.NoError(t, err)
		assert.Equal(t, "budi_findbyusername", user.Username)
		assert.Equal(t, "budi_findbyusername@example.com", user.Email)
	})
	t.Run("failed - user not found", func(t *testing.T) {
		user, err := suite.repository.FindByUsername(ctx, "salah")
		assert.Error(t, err)
		assert.Empty(t, user)
		assert.Equal(t, ErrNotFound, err)
	})
}
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

// func (suite *UserRepositoryTestSuite) TearDownTest() {
// 	ctx := context.Background()
	
// 	_, err := suite.pgContainer.Pool.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
// 	suite.NoError(err)
// }
