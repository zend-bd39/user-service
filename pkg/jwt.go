package pkg

import (
	"errors"
	"log"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token is expired")
	ErrSigningMethod = errors.New("signing method failed")
)

type Claims struct {
	UserID int `json:"user_id"`
	Role string `json:"role"`
	TokenType string `json:"token_type"`
	jwtv5.RegisteredClaims
}

type JWTService struct {
	secret []byte
	accessTTL time.Duration
	refreshTTL time.Duration
}

func NewJWTService(sec string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secret: []byte(sec),
		accessTTL: accessTTL,
		refreshTTL: refreshTTL,
	}
}
func (j *JWTService) GenerateAccessToken(userID int, role string)(string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role: role,
		TokenType: "access",
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(j.accessTTL)),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) GenerateRefreshToken(userID int)(string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		TokenType: "refresh",
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(j.refreshTTL)),
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) VerifyToken(token string) (*Claims, error) {
	claims := &Claims{}
	jwtToken, err := jwtv5.ParseWithClaims(token, claims, func(t *jwtv5.Token) (any, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			log.Printf("[JWT SERVICE] error when assertion method: %s", t.Header["alg"])
			return nil, ErrSigningMethod
		}
		return j.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwtv5.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	extractedClaims, ok := jwtToken.Claims.(*Claims)
	if !ok || !jwtToken.Valid {
		return nil, ErrInvalidToken
	}
	return extractedClaims, nil
}