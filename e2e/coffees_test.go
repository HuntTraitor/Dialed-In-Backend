package e2e

import (
	"bytes"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/mocks"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestGetAllCoffees(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	authenticateBody := authenticateUser(t, "hunter@gmail.com", "password")
	token := authenticateBody["authentication_token"].(map[string]any)["token"].(string)
	createCoffee(t, token, *mocks.MockCoffee, []byte(mocks.MockCoffee.Info.Img))

	t.Run("Successfully gets list of coffees", func(t *testing.T) {
		// Send request to coffee endpoint and extract body
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
		requestHeaders := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		statusCode, _, respBody := get(t, requestURL, requestHeaders)

		// assert status codes are equal
		assert.Equal(t, http.StatusOK, statusCode)

		// assert body is correct
		coffees := respBody["coffees"].([]any)
		for _, coffee := range coffees {
			c := coffee.(map[string]any)

			assert.NotEmpty(t, c["id"].(float64))
			assert.NotEmpty(t, c["user_id"].(float64))
			assert.NotEmpty(t, c["created_at"].(string))
			assert.Equal(t, float64(1), c["version"].(float64))

			info := c["info"].(map[string]any)
			assert.NotEmpty(t, info["img"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.Name, info["name"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.Roaster, info["roaster"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.Region, info["region"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.Process, info["process"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.Description, info["description"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.OriginType, info["origin_type"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.Rating, int(info["rating"].(float64)))
			assert.Equal(t, mocks.MockCoffee.Info.RoastLevel, info["roast_level"].(string))
			assert.Equal(t, mocks.MockCoffee.Info.Cost, info["cost"].(float64))
			assert.Equal(t, mocks.MockCoffee.Info.Decaf, info["decaf"].(bool))
			actualNotes := make([]string, len(info["tasting_notes"].([]any)))
			for i, note := range info["tasting_notes"].([]any) {
				actualNotes[i] = note.(string)
			}
			assert.ElementsMatch(t, mocks.MockCoffee.Info.TastingNotes, actualNotes)
		}
	})

	t.Run("Fails to get coffees when not logged in", func(t *testing.T) {
		// Send request with no auth header
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
		statusCode, _, respBody := get(t, requestURL, nil)

		// Assert there is a failed response
		assert.Equal(t, http.StatusUnauthorized, statusCode)
		assert.Equal(t, "you must be authenticated to access this resource", respBody["error"].(string))
	})
}

func TestPostCoffee(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	resp := authenticateUser(t, "hunter@gmail.com", "password")
	token := resp["authentication_token"].(map[string]any)["token"].(string)

	longName := strings.Repeat("A", 510)
	longRoaster := strings.Repeat("A", 510)
	longRegion := strings.Repeat("B", 110)
	longProcess := strings.Repeat("p", 210)
	longDescription := strings.Repeat("C", 1010)
	longOriginType := strings.Repeat("o", 110)
	longTastingNote := []string{strings.Repeat("t", 110)}
	longTastingNotes := make([]string, 51)
	for i := range longTastingNotes {
		longTastingNotes[i] = fmt.Sprintf("note%d", i+1)
	}

	tests := []struct {
		name               string
		payload            map[string]any
		expectedStatusCode int
		expectedResponse   map[string]any
		expectedError      map[string]any
	}{
		{
			name: "Successfully posts a new coffee",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse: map[string]any{
				"id":      1,
				"user_id": 1,
				"info": map[string]any{
					"name":          mocks.MockCoffee.Info.Name,
					"roaster":       mocks.MockCoffee.Info.Roaster,
					"region":        mocks.MockCoffee.Info.Region,
					"process":       mocks.MockCoffee.Info.Process,
					"description":   mocks.MockCoffee.Info.Description,
					"origin_type":   mocks.MockCoffee.Info.OriginType,
					"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
					"rating":        mocks.MockCoffee.Info.Rating,
					"roast_level":   mocks.MockCoffee.Info.RoastLevel,
					"cost":          mocks.MockCoffee.Info.Cost,
					"decaf":         mocks.MockCoffee.Info.Decaf,
					"img":           mocks.MockCoffee.Info.Img,
					"created_at":    "2025-02-01T03:59:07Z",
					"version":       1,
				},
			},
			expectedError: nil,
		},
		{
			name: "Coffee name too long returns an error",
			payload: map[string]any{
				"name":          longName,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"name": "must not be more than 500 bytes long",
			},
		},
		{
			name: "Coffee region too long returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        longRegion,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"region": "must not be more than 100 bytes long",
			},
		},
		{
			name: "Coffee description too long returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   longDescription,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"description": "must not be more than 1000 bytes long",
			},
		},
		{
			name: "Coffee process name too long returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       longProcess,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"process": "must not be more than 200 bytes long",
			},
		},
		{
			name: "Coffee roaster too long return an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       longRoaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"roaster": "must not be more than 200 bytes long",
			},
		},
		{
			name: "Coffee Origin Type too long returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   longOriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"origin_type": "must not be more than 100 bytes long",
			},
		},
		{
			name: "Coffee Tasting Notes amount too long returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": longTastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"tasting_notes": "must not contain more than 50 entries",
			},
		},
		{
			name: "Coffee Testing Note length too long returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": longTastingNote,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"tasting_notes[0]": "must not be more than 100 bytes long",
			},
		},
		{
			name: "Coffee rating less than zero returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        -1,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"rating": "must be between 0 and 5",
			},
		},
		{
			name: "Coffee rating more than 5 returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        6,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          mocks.MockCoffee.Info.Cost,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"rating": "must be between 0 and 5",
			},
		},
		{
			name: "Coffee cost less than zero returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          -1.0,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"cost": "must not be less than 0",
			},
		},
		{
			name: "Coffee cost more than 1,000,000 returns an error",
			payload: map[string]any{
				"name":          mocks.MockCoffee.Info.Name,
				"roaster":       mocks.MockCoffee.Info.Roaster,
				"region":        mocks.MockCoffee.Info.Region,
				"process":       mocks.MockCoffee.Info.Process,
				"description":   mocks.MockCoffee.Info.Description,
				"origin_type":   mocks.MockCoffee.Info.OriginType,
				"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
				"rating":        mocks.MockCoffee.Info.Rating,
				"roast_level":   mocks.MockCoffee.Info.RoastLevel,
				"cost":          1_000_001.00,
				"decaf":         mocks.MockCoffee.Info.Decaf,
				"img":           mocks.MockCoffee.Info.Img,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"cost": "must not be more than 1,000,000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)

			var b bytes.Buffer
			writer := multipart.NewWriter(&b)

			// Add form fields
			writer.WriteField("name", tt.payload["name"].(string))
			writer.WriteField("roaster", tt.payload["roaster"].(string))
			writer.WriteField("region", tt.payload["region"].(string))
			writer.WriteField("process", tt.payload["process"].(string))
			writer.WriteField("description", tt.payload["description"].(string))
			writer.WriteField("origin_type", tt.payload["origin_type"].(string))
			writer.WriteField("rating", strconv.Itoa(tt.payload["rating"].(int)))
			writer.WriteField("roast_level", tt.payload["roast_level"].(string))
			writer.WriteField("decaf", strconv.FormatBool(tt.payload["decaf"].(bool)))
			writer.WriteField("cost", strconv.FormatFloat(tt.payload["cost"].(float64), 'f', 2, 64))

			// Add tasting notes form field
			notes := tt.payload["tasting_notes"].([]string)
			for _, note := range notes {
				writer.WriteField("tasting_notes", note)
			}
			// Add image as a form file
			fileWriter, err := writer.CreateFormFile("img", "mock.jpg")
			if err != nil {
				t.Fatalf("failed to create form file: %v", err)
			}

			fileWriter.Write([]byte(tt.payload["img"].(string)))
			if err != nil {
				t.Fatalf("failed to write image data: %v", err)
			}

			headers := map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
				"Content-Type":  writer.FormDataContentType(),
			}

			writer.Close()
			// Send the request and get the response
			statusCode, _, returnedBody := post(t, requestURL, &b, headers)

			assert.Equal(t, tt.expectedStatusCode, statusCode)

			if tt.expectedError == nil {
				returnedCoffee := returnedBody["coffee"].(map[string]any)
				info := returnedCoffee["info"].(map[string]any)
				expectedInfo := tt.expectedResponse["info"].(map[string]any)

				assert.Equal(t, expectedInfo["name"], info["name"])
				assert.Equal(t, expectedInfo["roaster"], info["roaster"])
				assert.Equal(t, expectedInfo["region"], info["region"])
				assert.Equal(t, expectedInfo["process"], info["process"])
				assert.Equal(t, expectedInfo["description"], info["description"])
				assert.Equal(t, expectedInfo["origin_type"], info["origin_type"])
				assert.Equal(t, expectedInfo["cost"], info["cost"])
				assert.Equal(t, expectedInfo["roast_level"], info["roast_level"])
				assert.Equal(t, float64(expectedInfo["rating"].(int)), info["rating"])
				assert.Equal(t, expectedInfo["decaf"], info["decaf"])
				assert.ElementsMatch(t, expectedInfo["tasting_notes"], info["tasting_notes"])

				assert.NotEmpty(t, returnedCoffee["version"])
				assert.NotEmpty(t, info["img"])
				assert.NotEmpty(t, returnedCoffee["id"])
				assert.NotEmpty(t, returnedCoffee["user_id"])
				assert.NotEmpty(t, returnedCoffee["created_at"])
			} else {
				returnedError := returnedBody["error"].(map[string]any)
				assert.Equal(t, tt.expectedError, returnedError)
			}
		})
	}

	t.Run("Creating minimal coffee omits all appropriate fields", func(t *testing.T) {
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		writer.WriteField("name", "Test Name")

		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
			"Content-Type":  writer.FormDataContentType(),
		}
		writer.Close()

		statusCode, _, returnedBody := post(t, requestURL, &b, headers)
		assert.Equal(t, 201, statusCode)
		returnedCoffee := returnedBody["coffee"].(map[string]any)
		info := returnedCoffee["info"].(map[string]any)
		assert.Equal(t, "Test Name", info["name"])
		assert.Equal(t, false, info["decaf"])
		assert.Empty(t, info["roaster"])
		assert.Empty(t, info["region"])
		assert.Empty(t, info["process"])
		assert.Empty(t, info["description"])
		assert.Empty(t, info["origin_type"])
		assert.Empty(t, info["cost"])
		assert.Empty(t, info["roast_level"])
		assert.Empty(t, info["rating"])
		assert.Empty(t, info["img"])
		assert.Empty(t, info["tasting_notes"])
	})

	t.Run("Empty params returns an error", func(t *testing.T) {
		expectedError := map[string]any{
			"name": "must be provided",
		}

		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)
		writer.Close()

		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
			"Content-Type":  writer.FormDataContentType(),
		}

		// Send the request and get the response
		statusCode, _, returnedBody := post(t, requestURL, &b, headers)

		assert.Equal(t, http.StatusUnprocessableEntity, statusCode)

		returnedError := returnedBody["error"].(map[string]any)
		assert.Equal(t, expectedError, returnedError)
	})

	t.Run("Unauthenticated call response when an error", func(t *testing.T) {

		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)

		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		// Add form fields
		writer.WriteField("name", "test")
		writer.WriteField("region", "test")
		writer.WriteField("process", "test")
		writer.WriteField("description", "test")

		// Add image as a form file
		fileWriter, err := writer.CreateFormFile("img", "Test Image")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		fileWriter.Write([]byte("test")) // Write mock image data

		headers := map[string]string{
			"Content-Type": writer.FormDataContentType(),
		}

		writer.Close()

		// Send the request and get the response
		statusCode, _, _ := post(t, requestURL, &b, headers)

		assert.Equal(t, http.StatusUnauthorized, statusCode)
	})
}

func TestUpdateCoffee(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	resp := authenticateUser(t, "hunter@gmail.com", "password")
	token := resp["authentication_token"].(map[string]any)["token"].(string)

	tests := []struct {
		name               string
		payload            map[string]any
		expectedStatusCode int
		expectedResponse   map[string]any
		expectedError      map[string]any
	}{
		{
			name: "Successfully updates a coffee",
			payload: map[string]any{
				"name":          "Updated Name",
				"roaster":       "Updated Roaster",
				"region":        "Updated Region",
				"process":       "Updated Process",
				"description":   "Updated Description",
				"origin_type":   "Updated Origin Type",
				"cost":          100.00,
				"roast_level":   "Updated Roast Level",
				"rating":        2,
				"decaf":         true,
				"tasting_notes": []string{"chocolate", "caramel"},
				"img":           []byte("Updated Image"),
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]any{
				"info": map[string]any{
					"name":          "Updated Name",
					"roaster":       "Updated Roaster",
					"region":        "Updated Region",
					"process":       "Updated Process",
					"description":   "Updated Description",
					"origin_type":   "Updated Origin Type",
					"cost":          100.00,
					"roast_level":   "Updated Roast Level",
					"rating":        2,
					"decaf":         true,
					"tasting_notes": []string{"chocolate", "caramel"},
					"img":           []byte("Updated Image"),
				},
			},
			expectedError: nil,
		},
		{
			name: "Successfully Partially updates a coffee",
			payload: map[string]any{
				"name": "Updated Name",
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]any{
				"info": map[string]any{
					"name":          "Updated Name",
					"roaster":       mocks.MockCoffee.Info.Roaster,
					"region":        mocks.MockCoffee.Info.Region,
					"process":       mocks.MockCoffee.Info.Process,
					"description":   mocks.MockCoffee.Info.Description,
					"origin_type":   mocks.MockCoffee.Info.OriginType,
					"cost":          mocks.MockCoffee.Info.Cost,
					"roast_level":   mocks.MockCoffee.Info.RoastLevel,
					"rating":        mocks.MockCoffee.Info.Rating,
					"decaf":         mocks.MockCoffee.Info.Decaf,
					"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
					"img":           mocks.MockCoffee.Info.Img,
				},
			},
			expectedError: nil,
		},
		{
			name:               "Updating with no fields is still successful",
			payload:            map[string]any{},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]any{
				"info": map[string]any{
					"name":          mocks.MockCoffee.Info.Name,
					"roaster":       mocks.MockCoffee.Info.Roaster,
					"region":        mocks.MockCoffee.Info.Region,
					"process":       mocks.MockCoffee.Info.Process,
					"description":   mocks.MockCoffee.Info.Description,
					"origin_type":   mocks.MockCoffee.Info.OriginType,
					"cost":          mocks.MockCoffee.Info.Cost,
					"roast_level":   mocks.MockCoffee.Info.RoastLevel,
					"rating":        mocks.MockCoffee.Info.Rating,
					"decaf":         mocks.MockCoffee.Info.Decaf,
					"tasting_notes": mocks.MockCoffee.Info.TastingNotes,
					"img":           mocks.MockCoffee.Info.Img,
				},
			},
			expectedError: nil,
		},
		{
			name: "Update with an unknown field returns an error",
			payload: map[string]any{
				"random_field": "unknown",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
			expectedError: map[string]any{
				"error": "body contains unknown key \"random_field\"",
			},
		},
		{
			name: "Update with a known AND unknown field returns an error",
			payload: map[string]any{
				"name":         "Updated Name",
				"region":       "Updated Region",
				"process":      "Updated Process",
				"img":          []byte(""),
				"description":  "Updated Description",
				"random_field": "unknown",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   nil,
			expectedError: map[string]any{
				"error": "body contains unknown key \"random_field\"",
			},
		},
	}

	// TEST TABLE CASES
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// post a new coffee
			postedCoffee := createCoffee(t, token, *mocks.MockCoffee, []byte("Test Image"))
			postedCoffeeID := int(postedCoffee["coffee"].(map[string]any)["id"].(float64))

			// send a patch request to that url
			var b bytes.Buffer
			writer := multipart.NewWriter(&b)

			for key, value := range tt.payload {
				if key == "img" {
					fileWriter, err := writer.CreateFormFile("img", "Test Image")
					if err != nil {
						t.Fatalf("failed to create form file: %v", err)
					}
					fileWriter.Write(tt.payload["img"].([]byte)) // Write mock image data
				} else if key == "tasting_notes" {
					notes := value.([]string)
					for _, note := range notes {
						writer.WriteField("tasting_notes", note)
					}
				} else {
					writer.WriteField(key, fmt.Sprintf("%v", value))
				}
			}

			headers := map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
				"Content-Type":  writer.FormDataContentType(),
			}
			writer.Close()

			patchURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, postedCoffeeID)
			statusCode, _, returnedBody := patch(t, patchURL, &b, headers)
			assert.Equal(t, tt.expectedStatusCode, statusCode)

			if tt.expectedResponse != nil {
				coffeeBody := returnedBody["coffee"].(map[string]any)
				info := coffeeBody["info"].(map[string]any)
				expectedInfo := tt.expectedResponse["info"].(map[string]any)
				for k, v := range expectedInfo {
					switch k {
					case "img":
						assert.NotEmpty(t, info["img"])
					case "tasting_notes":
						assert.ElementsMatch(t, v, info["tasting_notes"])
					case "rating":
						assert.Equal(t, float64(v.(int)), info["rating"])
					default:
						assert.Equal(t, v, info[k])
					}
				}
			}
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, returnedBody)
			}

		})
	}

	t.Run("Updating an item that does not exist returns an error", func(t *testing.T) {
		// send a patch request to that url
		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)
		writer.Close()
		patchURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, 47834957)
		statusCode, _, _ := patch(t, patchURL, &b, headers)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("Updating an item that the user does not own returns an error", func(t *testing.T) {

		// create and log into new user
		createUser(t)
		res := authenticateUser(t, "test@example.com", "password")
		newToken := res["authentication_token"].(map[string]any)["token"].(string)

		// create coffee as that user
		returnedBody := createCoffee(t, newToken, *mocks.MockCoffee, []byte("Test Image"))
		newPostedCoffeeID := int(returnedBody["coffee"].(map[string]any)["id"].(float64))

		// send a patch request to that url
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		// send patch request as the old user on the new record
		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
			"Content-Type":  writer.FormDataContentType(),
		}
		writer.Close()

		patchURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, newPostedCoffeeID)
		statusCode, _, _ := patch(t, patchURL, &b, headers)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("Unauthenticated user updating a coffee returns an error", func(t *testing.T) {
		patchURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, 1)
		// send a patch request to that url
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		// send patch request as the old user on the new record
		headers := map[string]string{
			"Content-Type": writer.FormDataContentType(),
		}
		writer.Close()
		statusCode, _, _ := patch(t, patchURL, &b, headers)
		assert.Equal(t, http.StatusUnauthorized, statusCode)
	})

	t.Run("Sending a patch request that is not a multi part form returns an error", func(t *testing.T) {
		patchURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, 1)
		// send a patch request to that url
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		// send patch request as the old user on the new record
		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		writer.Close()
		statusCode, _, returnedBody := patch(t, patchURL, &b, headers)
		assert.Equal(t, http.StatusBadRequest, statusCode)
		assert.Equal(t, map[string]any{
			"error": "content type must be multipart/form-data",
		}, returnedBody)
	})
}

func TestDeleteCoffee(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	resp := authenticateUser(t, "hunter@gmail.com", "password")
	token := resp["authentication_token"].(map[string]any)["token"].(string)

	t.Run("Successfully deletes a coffee", func(t *testing.T) {
		// post a new coffee and extract ID
		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}

		// post a new coffee
		returnedBody := createCoffee(t, token, *mocks.MockCoffee, []byte("Test Image"))
		postedCoffeeID := int(returnedBody["coffee"].(map[string]any)["id"].(float64))

		// get request to that new coffee returns 200
		getURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, postedCoffeeID)
		statusCode, _, _ := get(t, getURL, headers)
		assert.Equal(t, http.StatusOK, statusCode)

		// delete the coffee
		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, postedCoffeeID)
		statusCode, _, returnedBody = delete(t, deleteURL, headers)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.NotEmpty(t, returnedBody["message"])

		// get request of the deleted coffee returns 404
		statusCode, _, _ = get(t, getURL, headers)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("Deleting a coffee that does not exist returns an error", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}

		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, 10000)
		statusCode, _, _ := delete(t, deleteURL, headers)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("Deleting a coffee that the user does not own returns an error", func(t *testing.T) {
		// create a log in as a new user
		createUser(t)
		res := authenticateUser(t, "test@example.com", "password")
		newToken := res["authentication_token"].(map[string]any)["token"].(string)

		// post a request as that user
		returnedBody := createCoffee(t, newToken, *mocks.MockCoffee, []byte("Test Image"))
		newPostedCoffeeID := int(returnedBody["coffee"].(map[string]any)["id"].(float64))

		// send delete request as the old user on the new record
		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, newPostedCoffeeID)
		statusCode, _, _ := delete(t, deleteURL, headers)
		assert.Equal(t, http.StatusNotFound, statusCode)
	})

	t.Run("Deleting a coffee when the user is not authenticated returns an error", func(t *testing.T) {
		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, 1)
		statusCode, _, _ := delete(t, deleteURL, nil)
		assert.Equal(t, http.StatusUnauthorized, statusCode)
	})
}
