package main

//
//import (
//	"bytes"
//	"fmt"
//	"github.com/hunttraitor/dialed-in-backend/internal/assert"
//	"net/http"
//	"testing"
//)
//
//func TestCreateUser(t *testing.T) {
//	app := newTestApplication(t)
//	ts := newTestServer(t, app.routes())
//	defer ts.Close()
//
//	tests := []struct {
//		name               string
//		payload            string
//		expectedStatusCode int
//		expectedWrapper    string
//		expectedResponse   string
//	}{
//		{
//			name: "Successfully creates new user",
//			payload: `{
//				"name":     "Test User",
//				"email":    "test@example.com",
//				"password": "password",
//			}`,
//			expectedStatusCode: http.StatusCreated,
//			expectedWrapper:    "user",
//			expectedResponse: `{
//				"user": {
//						"id": 1,
//						"created_at": "2024-11-14T21:46:09Z",
//						"name": "Test User",
//						"email": "test@example.com",
//						"activated": false
//				}
//			}`,
//		},
//		{
//			name: "Creates a duplicate user",
//			payload: `{
//				"name":     "Test User",
//				"email":    "test@example.com",
//				"password": "password",
//			}`,
//			expectedStatusCode: http.StatusUnprocessableEntity,
//			expectedWrapper:    "error",
//			expectedResponse: `{
//				"error": {
//						"email": "a user with this email address already exists"
//				}
//			}`,
//		},
//		{
//			name:               "No body provided",
//			payload:            ``,
//			expectedStatusCode: http.StatusUnprocessableEntity,
//			expectedWrapper:    "error",
//			expectedResponse: `{
//				"error": {
//						"email": "must be provided",
//						"name": "must be provided",
//						"password": "must be provided"
//				}
//			}`,
//		},
//		{
//			name: "Inputting a bad email",
//			payload: `{
//				"name":     "Test User",
//				"email":    "testexample.com",
//				"password": "password",
//			}`,
//			expectedStatusCode: http.StatusUnprocessableEntity,
//			expectedWrapper:    "error",
//			expectedResponse: `{
//				"error": {
//						"email": "must be a valid email address"
//				}
//			}`,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//
//			// Marshall input data to string
//			//inputData, err := json.Marshal(tt.payload)
//			//if err != nil {
//			//	t.Fatal(err)
//			//}
//
//			inputData := []byte(tt.payload)
//
//			// Send request and expect the response to be OK
//			code, _, body := ts.post(t, "/v1/users", bytes.NewBuffer(inputData))
//			fmt.Println(body)
//			assert.Equal(t, tt.expectedStatusCode, code)
//
//			expectedBody := assert.UnmarshalAndUnwrap(t, tt.expectedWrapper, tt.expectedResponse)
//			actualBody := assert.UnmarshalAndUnwrap(t, tt.expectedWrapper, body)
//
//			for k, v := range expectedBody {
//				switch k {
//				case "id":
//					assert.NotNil(t, actualBody[k])
//				case "created_at":
//					assert.NotEmpty(t, actualBody[k].(string))
//				default:
//					assert.Equal(t, v, actualBody[k])
//				}
//			}
//		})
//	}
//}
