package data

import (
	"fmt"
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
			m := MethodModel{db}

			methods, err := m.GetAll()
			if err != nil {
				t.Fatal(err)
			}

			for i := range methods {
				fmt.Println(methods[i].CreatedAt)
				assert.Equal(t, tt.expectedResult[i].Name, methods[i].Name)
				assert.Equal(t, tt.expectedResult[i].Img, methods[i].Img)
			}
		})
	}

}
