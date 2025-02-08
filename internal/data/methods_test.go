package data

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAll(t *testing.T) {
	db := newTestDB(t)
	s3 := newTestS3(t)
	tests := []struct {
		name           string
		expectedResult []Method
	}{
		{
			name: "Successfully gets list of methods",
			expectedResult: []Method{
				{
					Name: "Pour Over",
					Img:  "https://example.com/images/pour_over.png",
				},
				{
					Name: "Hario Switch",
					Img:  "https://example.com/images/hario_switch.png",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MethodModel{db, &s3}

			methods, err := m.GetAll()
			if err != nil {
				t.Fatal(err)
			}

			for i := range methods {
				assert.Equal(t, tt.expectedResult[i].Name, methods[i].Name)
				assert.NotEmpty(t, methods[i].Img)
			}
		})
	}

}
