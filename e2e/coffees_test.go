package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
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
	insertedCoffee := data.Coffee{
		Name:        "Test Coffee",
		Region:      "Test Region",
		Process:     "Test Process",
		Description: "Test Description",
		Img:         "Test Image",
	}
	createCoffee(t, token, insertedCoffee, []byte(insertedCoffee.Img))

	t.Run("Successfully gets list of coffees", func(t *testing.T) {
		// Send request to coffee endpoint and extract body
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
		requestHeaders := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		statusCode, _, respBody := get(t, requestURL, requestHeaders)
		var body map[string]any
		err = json.Unmarshal([]byte(respBody), &body)
		if err != nil {
			t.Fatal(err)
		}

		// assert status codes are equal
		assert.Equal(t, http.StatusOK, statusCode)

		// assert body is correct
		coffees := body["coffees"].([]any)
		for _, coffee := range coffees {
			c := coffee.(map[string]any)
			assert.NotEmpty(t, c["id"].(float64))
			assert.NotEmpty(t, c["user_id"].(float64))
			assert.NotEmpty(t, c["created_at"].(string))
			assert.NotEmpty(t, c["img"].(string))
			assert.Equal(t, insertedCoffee.Name, c["name"].(string))
			assert.Equal(t, insertedCoffee.Region, c["region"].(string))
			assert.Equal(t, insertedCoffee.Process, c["process"].(string))
			assert.Equal(t, insertedCoffee.Description, c["description"].(string))
			assert.Equal(t, float64(1), c["version"].(float64))
		}
	})

	t.Run("Fails to get coffees when not logged in", func(t *testing.T) {
		// Send request with no auth header
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
		statusCode, _, respBody := get(t, requestURL, nil)
		var body map[string]any
		err = json.Unmarshal([]byte(respBody), &body)
		if err != nil {
			t.Fatal(err)
		}

		// Assert there is a failed response
		assert.Equal(t, http.StatusUnauthorized, statusCode)
		assert.Equal(t, "you must be authenticated to access this resource", body["error"].(string))
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
	longRegion := strings.Repeat("B", 110)
	longProcess := strings.Repeat("p", 210)
	longDescription := strings.Repeat("C", 1010)

	mockCoffee := struct {
		name        string
		region      string
		process     string
		description string
		image       []byte
	}{
		name:        "Mock Coffee",
		region:      "Mock Region",
		process:     "Mock Process",
		description: "Mock Description",
		image:       []byte("Mock Image"),
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
				"name":        mockCoffee.name,
				"region":      mockCoffee.region,
				"process":     mockCoffee.process,
				"description": mockCoffee.description,
				"image":       mockCoffee.image,
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse: map[string]any{
				"id":          1,
				"user_id":     1,
				"name":        mockCoffee.name,
				"region":      mockCoffee.region,
				"process":     mockCoffee.process,
				"description": mockCoffee.description,
				"img":         "",
				"created_at":  "2025-02-01T03:59:07Z",
				"version":     1,
			},
			expectedError: nil,
		},
		{
			name: "Coffee name too long returns an error",
			payload: map[string]any{
				"name":        longName,
				"region":      mockCoffee.region,
				"process":     mockCoffee.process,
				"description": mockCoffee.description,
				"image":       mockCoffee.image,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"name": "must not be more than 500 bytes long",
			},
		},
		{
			name: "Coffee region too long returns an error",
			payload: map[string]any{
				"name":        mockCoffee.name,
				"region":      longRegion,
				"process":     mockCoffee.process,
				"description": mockCoffee.description,
				"image":       mockCoffee.image,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"region": "must not be more than 100 bytes long",
			},
		},
		{
			name: "Coffee description too long returns an error",
			payload: map[string]any{
				"name":        mockCoffee.name,
				"region":      mockCoffee.region,
				"process":     mockCoffee.process,
				"description": longDescription,
				"image":       mockCoffee.image,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"description": "must not be more than 1000 bytes long",
			},
		},
		{
			name: "Coffee process name too long returns an error",
			payload: map[string]any{
				"name":        mockCoffee.name,
				"region":      mockCoffee.region,
				"process":     longProcess,
				"description": mockCoffee.description,
				"image":       mockCoffee.image,
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"process": "must not be more than 200 bytes long",
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
			writer.WriteField("region", tt.payload["region"].(string))
			writer.WriteField("process", tt.payload["process"].(string))
			writer.WriteField("description", tt.payload["description"].(string))

			// Add image as a form file
			fileWriter, err := writer.CreateFormFile("img", "Test Image")
			if err != nil {
				t.Fatalf("failed to create form file: %v", err)
			}
			fileWriter.Write(tt.payload["image"].([]byte)) // Write mock image data

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
				assert.Equal(t, tt.expectedResponse["name"], returnedCoffee["name"])
				assert.Equal(t, tt.expectedResponse["region"], returnedCoffee["region"])
				assert.Equal(t, tt.expectedResponse["process"], returnedCoffee["process"])
				assert.Equal(t, tt.expectedResponse["description"], returnedCoffee["description"])
				assert.NotEmpty(t, returnedCoffee["version"])
				assert.NotEmpty(t, returnedCoffee["id"])
				assert.NotEmpty(t, returnedCoffee["user_id"])
				assert.NotEmpty(t, returnedCoffee["created_at"])
				assert.NotEmpty(t, returnedCoffee["img"])
			} else {
				returnedError := returnedBody["error"].(map[string]any)
				assert.Equal(t, tt.expectedError, returnedError)
			}
		})
	}

	t.Run("Empty params returns an error", func(t *testing.T) {
		expectedError := map[string]any{
			"description": "must be provided",
			"name":        "must be provided",
			"process":     "must be provided",
			"region":      "must be provided",
			"img":         "must be provided",
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

	mockCoffee := data.Coffee{
		Name:        "Test Coffee",
		Region:      "Test Region",
		Process:     "Test Process",
		Description: "Test Description",
		Img:         "Test Image",
	}

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
				"name":        "Updated Name",
				"region":      "Updated Region",
				"process":     "Updated Process",
				"description": "Updated Description",
				"img":         []byte("Updated Image"),
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]any{
				"name":        "Updated Name",
				"region":      "Updated Region",
				"process":     "Updated Process",
				"img":         "",
				"description": "Updated Description",
				"version":     float64(2),
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
				"name":        "Updated Name",
				"region":      "Test Region",
				"process":     "Test Process",
				"img":         "",
				"description": "Test Description",
				"version":     float64(2),
			},
			expectedError: nil,
		},
		{
			name:               "Updating with no fields is still successful",
			payload:            map[string]any{},
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]any{
				"name":        "Test Coffee",
				"region":      "Test Region",
				"process":     "Test Process",
				"img":         "",
				"description": "Test Description",
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
			postedCoffee := createCoffee(t, token, mockCoffee, []byte("Test Image"))
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
				for k, v := range tt.expectedResponse {
					switch k {
					case "img":
						assert.NotEmpty(t, coffeeBody["img"])
					default:
						assert.Equal(t, v, coffeeBody[k])
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
		returnedBody := createCoffee(t, newToken, mockCoffee, []byte("Test Image"))
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

//func TestDeleteCoffee(t *testing.T) {
//	cleanup, _, err := LaunchTestProgram(port)
//	if err != nil {
//		t.Fatalf("failed to launch test program: %v", err)
//	}
//	t.Cleanup(cleanup)
//
//	resp := authenticateUser(t, "hunter@gmail.com", "password")
//	token := resp["authentication_token"].(map[string]any)["token"].(string)
//
//	mockPostCoffeePayload := `{
//    "name": "Blueberry Boom",
//    "region": "Ethiopia",
//    "img": "https://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
//    "description": "This is a delicious blueberry coffee :)"
//	}`
//
//	t.Run("Successfully deletes a coffee", func(t *testing.T) {
//		// post a new coffee and extract ID
//		headers := map[string]string{
//			"Authorization": fmt.Sprintf("Bearer %s", token),
//		}
//
//		// post the new coffee
//		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
//		statusCode, _, returnedBody := post(t, requestURL, strings.NewReader(mockPostCoffeePayload), headers)
//		assert.Equal(t, http.StatusCreated, statusCode)
//		postedCoffeeID := int(returnedBody["coffee"].(map[string]any)["id"].(float64))
//
//		// get request to that new coffee returns 200
//		getURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, postedCoffeeID)
//		statusCode, _, _ = get(t, getURL, headers)
//		assert.Equal(t, http.StatusOK, statusCode)
//
//		// delete the coffee
//		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, postedCoffeeID)
//		statusCode, _, returnedBody = delete(t, deleteURL, headers)
//		assert.Equal(t, http.StatusOK, statusCode)
//		assert.NotEmpty(t, returnedBody["message"])
//
//		// get request of the deleted coffee returns 404
//		statusCode, _, _ = get(t, getURL, headers)
//		assert.Equal(t, http.StatusNotFound, statusCode)
//	})
//
//	t.Run("Deleting a coffee that does not exist returns an error", func(t *testing.T) {
//		headers := map[string]string{
//			"Authorization": fmt.Sprintf("Bearer %s", token),
//		}
//
//		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, 10000)
//		statusCode, _, _ := delete(t, deleteURL, headers)
//		assert.Equal(t, http.StatusNotFound, statusCode)
//	})
//
//	t.Run("Deleting a coffee that the user does not own returns an error", func(t *testing.T) {
//		// create a log in as a new user
//		createUser(t)
//		res := authenticateUser(t, "test@example.com", "password")
//		newToken := res["authentication_token"].(map[string]any)["token"].(string)
//
//		// post a request as that user
//		newUserHeaders := map[string]string{
//			"Authorization": fmt.Sprintf("Bearer %s", newToken),
//		}
//		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
//		statusCode, _, returnedBody := post(t, requestURL, strings.NewReader(mockPostCoffeePayload), newUserHeaders)
//		assert.Equal(t, http.StatusCreated, statusCode)
//		newPostedCoffeeID := int(returnedBody["coffee"].(map[string]any)["id"].(float64))
//
//		// send delete request as the old user on the new record
//		headers := map[string]string{
//			"Authorization": fmt.Sprintf("Bearer %s", token),
//		}
//		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, newPostedCoffeeID)
//		statusCode, _, _ = delete(t, deleteURL, headers)
//		assert.Equal(t, http.StatusNotFound, statusCode)
//	})
//
//	t.Run("Deleting a coffee when the user is not authenticated returns an error", func(t *testing.T) {
//		deleteURL := fmt.Sprintf("http://localhost:%d/v1/coffees/%d", 3001, 1)
//		statusCode, _, _ := delete(t, deleteURL, nil)
//		assert.Equal(t, http.StatusUnauthorized, statusCode)
//	})
//}
