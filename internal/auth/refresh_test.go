package auth_test

import (
	"go-api-example/internal/auth"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefresh_Create(t *testing.T) {
	rt := auth.NewRefreshToken()

	res := rt.Create()

	assert.NotEmpty(t, res)
}
