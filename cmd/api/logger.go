package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
)

const maxLoggedBodyBytes = 8_000

type sanitizedMultipartBody struct {
	Type   string                              `json:"type"`
	Fields map[string][]string                 `json:"fields"`
	Files  map[string][]sanitizedMultipartFile `json:"files"`
}

type sanitizedMultipartFile struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	SizeBytes   int64  `json:"size_bytes"`
	Redacted    bool   `json:"redacted"`
}

func bodyForLog(raw []byte, contentType string) any {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}

	mediaType, _, _ := mime.ParseMediaType(contentType)

	switch {
	case mediaType == "multipart/form-data":
		return sanitizeMultipartBody(raw, contentType)

	case mediaType == "application/json" || json.Valid(raw):
		return jsonForLog(raw)

	default:
		return stringForLog(raw)
	}
}

func jsonForLog(raw []byte) any {
	var v any

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	if err := dec.Decode(&v); err != nil {
		return stringForLog(raw)
	}

	return v
}

func stringForLog(raw []byte) string {
	s := string(raw)

	if len(s) > maxLoggedBodyBytes {
		return s[:maxLoggedBodyBytes] + "...(truncated)"
	}

	return s
}

// sanitizeMultipartBody reads the body and removes huge image payloads from the multipart body
func sanitizeMultipartBody(raw []byte, contentType string) sanitizedMultipartBody {
	body := sanitizedMultipartBody{
		Type:   "multipart/form-data",
		Fields: map[string][]string{},
		Files:  map[string][]sanitizedMultipartFile{},
	}

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		body.Fields["_error"] = []string{"failed to parse content type: " + err.Error()}
		return body
	}

	boundary := params["boundary"]
	if boundary == "" {
		body.Fields["_error"] = []string{"missing multipart boundary"}
		return body
	}

	reader := multipart.NewReader(bytes.NewReader(raw), boundary)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			body.Fields["_error"] = []string{"failed to read multipart part: " + err.Error()}
			break
		}

		fieldName := part.FormName()
		if fieldName == "" {
			continue
		}

		filename := part.FileName()
		if filename == "" {
			value, err := io.ReadAll(io.LimitReader(part, maxLoggedBodyBytes+1))
			if err != nil {
				body.Fields[fieldName] = append(body.Fields[fieldName], "[error reading field: "+err.Error()+"]")
				continue
			}

			body.Fields[fieldName] = append(body.Fields[fieldName], stringForLog(value))
			continue
		}

		size, _ := io.Copy(io.Discard, part)

		body.Files[fieldName] = append(body.Files[fieldName], sanitizedMultipartFile{
			Filename:    filename,
			ContentType: part.Header.Get("Content-Type"),
			SizeBytes:   size,
			Redacted:    true,
		})
	}

	return body
}
