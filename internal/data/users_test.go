package data

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

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

func TestGetByEmail(t *testing.T) {
	db := newTestDB(t)
	m := UserModel{db}
	err := m.Insert(&User{
		Name:  "Test User",
		Email: "testuser@example.com",
		Password: password{
			hash: []byte("hash"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name         string
		email        string
		expectedErr  error
		expectedUser *User
	}{
		{
			name:  "Successfully retrieves user",
			email: "testuser@example.com",
			expectedUser: &User{
				ID:    1,
				Name:  "Test User",
				Email: "testuser@example.com",
				Password: password{
					hash: []byte("hash"),
				},
				Activated: false,
				Version:   1,
			},
			expectedErr: nil,
		},
		{
			name:         "User not found",
			email:        "notfound@example.com",
			expectedErr:  ErrRecordNotFound,
			expectedUser: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := m.GetByEmail(tt.email)
			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, tt.expectedUser.Name, user.Name)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				assert.Equal(t, tt.expectedUser.Password, user.Password)
				assert.Equal(t, tt.expectedUser.Activated, user.Activated)
				assert.Equal(t, tt.expectedUser.Version, user.Version)
				assert.NotEmpty(t, user.CreatedAt)
			}
		})
	}
}

func TestUserUpdate(t *testing.T) {
	db := newTestDB(t)
	m := UserModel{db}
	tests := []struct {
		name        string
		inputUser   *User
		updatedUser *User
		expectedErr error
	}{
		{
			name: "Successfully updates Name, Email, password Hash, Activated, Version for a user",
			inputUser: &User{
				Name:  "Test User",
				Email: "testuser@example.com",
				Password: password{
					hash: []byte("hash"),
				},
			},
			updatedUser: &User{
				Name:  "Test User Updated",
				Email: "testuserupdated@example.com",
				Password: password{
					hash: []byte("updated_hash"),
				},
				Activated: true,
				Version:   2,
			},
			expectedErr: nil,
		},
		{
			name: "Updating to duplicate email causes error",
			inputUser: &User{
				Name:  "Test User",
				Email: "testuser2@example.com",
				Password: password{
					hash: []byte("hash"),
				},
			},
			updatedUser: &User{
				Name:  "Test User Updated",
				Email: "testuserupdated@example.com",
				Password: password{
					hash: []byte("updated_hash"),
				},
				Activated: true,
				Version:   2,
			},
			expectedErr: ErrDuplicateEmail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inserts the input User
			err := m.Insert(tt.inputUser)
			if err != nil {
				t.Fatal(err)
			}

			// update the updated user to match input useres id, createdat, and version
			tt.updatedUser.ID = tt.inputUser.ID
			tt.updatedUser.CreatedAt = tt.inputUser.CreatedAt
			tt.updatedUser.Version = tt.inputUser.Version

			err = m.Update(tt.updatedUser)
			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				// getByEmail and check that the user is correctly updated
				user, err := m.GetByEmail(tt.updatedUser.Email)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.updatedUser.Name, user.Name)
				assert.Equal(t, tt.updatedUser.Email, user.Email)
				assert.Equal(t, tt.updatedUser.Password, user.Password)
				assert.Equal(t, tt.updatedUser.Activated, user.Activated)
				assert.Equal(t, tt.updatedUser.Version, user.Version)
			}
		})
	}
}

func TestGetForToken(t *testing.T) {
	db := newTestDB(t)
	tm := TokenModel{db}
	m := UserModel{db}
	// insert a user
	user := &User{
		Name:  "Test User",
		Email: "testuser@example.com",
		Password: password{
			hash: []byte("hash"),
		},
		Activated: false,
	}
	err := m.Insert(user)
	if err != nil {
		t.Fatal(err)
	}

	// give that user a token
	token, err := tm.New(user.ID, 5*time.Minute, "scope")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		inputToken   string
		inputScope   string
		expectedUser *User
		expectedErr  error
	}{
		{
			name:         "Successfully retrieves user from token",
			inputToken:   token.Plaintext,
			inputScope:   token.Scope,
			expectedUser: user,
			expectedErr:  nil,
		},
		{
			name:         "Wrong token inputted returns user not found",
			inputToken:   "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			inputScope:   token.Scope,
			expectedUser: nil,
			expectedErr:  ErrRecordNotFound,
		},
		{
			name:         "Wrong scope inputted returns user not found",
			inputToken:   token.Plaintext,
			inputScope:   "unknown",
			expectedUser: nil,
			expectedErr:  ErrRecordNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := m.GetForToken(tt.inputScope, tt.inputToken)
			assert.Equal(t, tt.expectedErr, err)
			if err == nil {
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}

	t.Run("Return ErrRecordNotFound when token is expired", func(t *testing.T) {
		// Set token to expire 1 hour in the past
		token, err := tm.New(user.ID, -1*time.Hour, "scope")
		if err != nil {
			t.Fatal(err)
		}
		//time.Sleep(time.Second)
		_, err = m.GetForToken("scope", token.Plaintext)
		assert.Equal(t, ErrRecordNotFound, err)
	})
}
