package model_test

import (
	"go-api-example/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserEvent_GetID(t *testing.T) {
	tests := []struct {
		name      string
		userEvent *model.UserEvent
		wantID    string
	}{
		{
			name: "success",
			userEvent: &model.UserEvent{
				ID:        1,
				Username:  "johndoe",
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
			wantID: "1-johndoe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.userEvent.GetID()

			assert.Equal(t, tt.wantID, id)
		})
	}

}
