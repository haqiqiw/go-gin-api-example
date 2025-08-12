package model

import "go-api-example/internal/auth"

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=8,max=64"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type LogoutRequest struct {
	Claims       *auth.JWTClaims `json:"claims"`
	RefreshToken string          `json:"refresh_token" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
