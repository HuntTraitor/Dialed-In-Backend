package data

import (
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/lib/pq"
)

func TestUserCreate(t *testing.T) {
	db := newTestDB(t)
	tests := []struct {
		name        string
		user        *User
		expectedErr error
	}{
		{
			name: "Successfully inserts user",
			user: &User{
				Name:  "Test User",
				Email: "testuser@example.com",
				Password: password{
					hash: []byte("hash"),
				},
				Activated: false,
			},
			expectedErr: nil,
		},
		{
			name: "Duplicate email for user",
			user: &User{
				Name:  "Test User",
				Email: "testuser@example.com",
				Password: password{
					hash: []byte("hash"),
				},
				Activated: false,
			},
			expectedErr: ErrDuplicateEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UserModel{db}

			err := m.Insert(tt.user)
			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.NotEmpty(t, tt.user.ID)
				assert.NotEmpty(t, tt.user.CreatedAt)
				assert.NotEmpty(t, tt.user.Version)
			}
		})
	}
}
