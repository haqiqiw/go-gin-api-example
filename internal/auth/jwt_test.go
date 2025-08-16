package auth_test

import (
	"go-api-example/internal/auth"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTToken_Create(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "sucesss",
			userID:    "1",
			secretKey: "dummy-secret",
			wantErr:   false,
		},
		{
			name:      "sucesss with empty secret",
			userID:    "1",
			secretKey: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt := auth.NewJWTToken(tt.secretKey, time.Second)
			token, err := jwt.Create(tt.userID)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.NotEmpty(t, token)
		})
	}
}

func TestJWTToken_Parse(t *testing.T) {
	validSecret := "valid-secret"
	invalidSecret := "invalid-secret"
	jwt := auth.NewJWTToken(validSecret, 10*time.Second)
	validToken, _ := jwt.Create("1")

	tests := []struct {
		name       string
		token      string
		wantUser   string
		wantErrMsg string
	}{
		{
			name:       "valid token",
			token:      validToken,
			wantUser:   "1",
			wantErrMsg: "",
		},
		{
			name: "invalid token",
			token: func() string {
				invToken, _ := auth.
					NewJWTToken(invalidSecret, 10*time.Second).
					Create("1")
				return invToken
			}(),
			wantErrMsg: "token signature is invalid: signature is invalid",
		},
		{
			name: "expired token",
			token: func() string {
				expToken, _ := auth.
					NewJWTToken(validSecret, 10*time.Millisecond).
					Create("1")
				// wait token expired
				time.Sleep(20 * time.Millisecond)
				return expToken
			}(),
			wantErrMsg: "token has invalid claims: token is expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwt.Parse(tt.token)
			log.Println(claims)
			log.Println(err)

			if tt.wantErrMsg != "" {
				assert.Nil(t, claims)
				assert.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				assert.Equal(t, tt.wantUser, claims.UserID)
				assert.Nil(t, err)
			}
		})
	}
}
