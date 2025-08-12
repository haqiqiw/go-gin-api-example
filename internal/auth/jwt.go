package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	PrefixRevokeKey = "revoke-jwt-token"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTToken interface {
	Create(userID string) (string, error)
	Parse(jwtToken string) (*JWTClaims, error)
}

type jwtToken struct {
	SecretKey string
}

func NewJWTToken(secretKey string) JWTToken {
	return &jwtToken{
		SecretKey: secretKey,
	}
}

func (j *jwtToken) Create(userID string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *jwtToken) Parse(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid jwt claims")
	}

	if claims.ExpiresAt.Time.UnixMilli() < time.Now().UnixMilli() {
		return nil, errors.New("expired jwt claims")
	}

	return claims, nil
}
