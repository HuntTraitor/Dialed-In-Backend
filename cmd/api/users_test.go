package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateUser(t *testing.T) {

	payload := map[string]string{"name": "Hunter Tratar", "email": "hunterrrisatratar@gmail.com", "password": "password"}
	jsonData, err := json.Marshal(payload)
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.post(t, "/v1/users", headers, bytes.NewBuffer(jsonData))
	fmt.Println(code)
	fmt.Println(body)
}
