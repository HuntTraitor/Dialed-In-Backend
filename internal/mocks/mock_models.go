package mocks

import "github.com/hunttraitor/dialed-in-backend/internal/data"

// NewMockModels returns models that are meant for mocking
func NewMockModels() data.Models {
	return data.Models{
		Users: MockUserModel{},
	}
}
