package pkg

import (
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type JwtServiceTestSuite struct {
	suite.Suite
	service *JWTService
}
func (s *JwtServiceTestSuite) SetupSuite() {
	s.service = NewJWTService("123", 1*time.Second, 2*time.Second)
}
func (s *JwtServiceTestSuite) TestJWTService() {
	t := s.T()
	t.Run("success - generate access token", func(t *testing.T) {
		access, err  := s.service.GenerateAccessToken(1, "user")
		assert.NoError(t, err)
		assert.NotEmpty(t, access)
	})
	t.Run("success - generate refresh token", func(t *testing.T) {
		refresh, err  := s.service.GenerateRefreshToken(1)
		assert.NoError(t, err)
		assert.NotEmpty(t, refresh)
	})
	t.Run("success - validate token", func(t *testing.T) {
		access, err := s.service.GenerateAccessToken(1, "user")
		assert.NoError(t,err)
		claims, err := s.service.VerifyToken(access)
		assert.NoError(t,err)
		assert.Equal(t, int(1), claims.UserID)
		assert.Equal(t, "user", claims.Role)
	})
	t.Run("failed - signing method", func(t *testing.T) {
		token := jwtv5.NewWithClaims(jwtv5.SigningMethodNone, Claims{
			UserID: 1,
			Role: "user",
			RegisteredClaims: jwtv5.RegisteredClaims{
				IssuedAt: jwtv5.NewNumericDate(time.Now()),
				ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(s.service.accessTTL)),
			},
		})
		tokenStr, err := token.SignedString(jwtv5.UnsafeAllowNoneSignatureType)
		assert.NoError(t, err)
		claims, err := s.service.VerifyToken(tokenStr)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrInvalidToken, err)
	})
	t.Run("failed - expired access token", func(t *testing.T) {
		access, err := s.service.GenerateAccessToken(1, "user")
		assert.NoError(t,err)
		time.Sleep(2 * time.Second)
		claims, err := s.service.VerifyToken(access)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrExpiredToken, err)
	})
	t.Run("failed - expired refresh token", func(t *testing.T) {
		refresh, err := s.service.GenerateRefreshToken(1)
		assert.NoError(t,err)
		time.Sleep(3 * time.Second)
		claims, err := s.service.VerifyToken(refresh)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrExpiredToken, err)
	})
}
func TestJwtServiceTestSuite(t *testing.T) {
	suite.Run(t, new(JwtServiceTestSuite))
}