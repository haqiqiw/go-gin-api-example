package auth

import (
	"time"

	"github.com/google/uuid"
)

const (
	PrefixRefreshKey = "refresh-token"
	RefreshTTL       = 7 * 24 * time.Hour
)

//go:generate mockery --name=RefreshToken --structname RefreshToken --outpkg=mocks --output=./../mocks
type RefreshToken interface {
	Create() string
}

type refreshToken struct{}

func NewRefreshToken() RefreshToken {
	return &refreshToken{}
}

func (r *refreshToken) Create() string {
	return uuid.NewString()
}
