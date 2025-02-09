package main

import (
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/s3"
	"io"
	"net/http"
	"strings"
)

func isValidFileType(file []byte) bool {
	fileType := http.DetectContentType(file)
	return strings.HasPrefix(fileType, "image/") // Only allow images
}

func (app *application) fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	_ = app.contextGetUser(r)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Retrieve the file from form data
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "Uploaded File: %s\n", handler.Filename)
	fmt.Fprintf(w, "File Size: %d\n", handler.Size)
	fmt.Fprintf(w, "MIME Header: %v\n", handler.Header)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !isValidFileType(fileBytes) {
		app.notFoundResponse(w, r)
		return
	}

	if err = s3.UploadToS3(*app.s3, fileBytes, handler.Filename, app.config.s3.bucket); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
