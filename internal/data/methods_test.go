package data

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAll(t *testing.T) {
	db := newTestDB(t)
	tests := []struct {
		name           string
		expectedResult []Method
	}{
		{
			name: "Successfully gets list of methods",
			expectedResult: []Method{
				{
					Name: "Pour Over",
				},
				{
					Name: "Hario Switch",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MethodModel{db}

			methods, err := m.GetAll()
			if err != nil {
				t.Fatal(err)
			}

			for i := range methods {
				assert.Equal(t, tt.expectedResult[i].Name, methods[i].Name)
			}
		})
	}

}
