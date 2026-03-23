package testutils

import (
	"strconv"
	"testing"

	"github.com/gavv/httpexpect"
)

type APIClient struct {
	BaseURL string
	Token   string
}

func (c *APIClient) expect(t *testing.T) *httpexpect.Expect {
	t.Helper()
	return httpexpect.New(t, c.BaseURL)
}

func (c *APIClient) GET(path string) *RequestBuilder {
	return &RequestBuilder{client: c, method: "GET", path: path}
}

func (c *APIClient) DELETE(path string) *RequestBuilder {
	return &RequestBuilder{client: c, method: "DELETE", path: path}
}

func (c *APIClient) POSTJSON(path string, body any) *RequestBuilder {
	return &RequestBuilder{client: c, method: "POST", path: path, json: body}
}

func (c *APIClient) PATCHJSON(path string, body any) *RequestBuilder {
	return &RequestBuilder{client: c, method: "PATCH", path: path, json: body}
}

func (c *APIClient) POSTMultipart(path string, form CoffeeForm) *RequestBuilder {
	return &RequestBuilder{client: c, method: "POST", path: path, form: &form}
}

func (c *APIClient) PATCHMultipart(path string, form CoffeeForm) *RequestBuilder {
	return &RequestBuilder{client: c, method: "PATCH", path: path, form: &form}
}

func (c *APIClient) PUTJSON(path string, body any) *RequestBuilder {
	return &RequestBuilder{client: c, method: "PUT", path: path, json: body}
}

func (c *APIClient) PATCHMultipartWithExtraFields(path string, form CoffeeForm, extraFields map[string]string) *RequestBuilder {
	return &RequestBuilder{
		client:      c,
		method:      "PATCH",
		path:        path,
		form:        &form,
		extraFields: extraFields,
	}
}

type RequestBuilder struct {
	client *APIClient
	method string
	path   string

	json        any
	form        *CoffeeForm
	extraFields map[string]string
}

func (r *RequestBuilder) Expect(t *testing.T) *httpexpect.Response {
	t.Helper()

	e := r.client.expect(t)

	var req *httpexpect.Request

	switch r.method {
	case "GET":
		req = e.GET(r.path)
	case "DELETE":
		req = e.DELETE(r.path)
	case "POST":
		req = e.POST(r.path)
	case "PATCH":
		req = e.PATCH(r.path)
	case "PUT":
		req = e.PUT(r.path)
	default:
		t.Fatalf("unsupported method: %s", r.method)
		return nil
	}

	if r.client.Token != "" {
		req = req.WithHeader("Authorization", "Bearer "+r.client.Token)
	}

	if r.json != nil {
		req = req.WithJSON(r.json)
	}

	if r.form != nil {
		req = applyCoffeeMultipart(req, *r.form, r.extraFields)
	}

	return req.Expect()
}

func applyCoffeeMultipart(req *httpexpect.Request, form CoffeeForm, extraFields map[string]string) *httpexpect.Request {
	mp := req.WithMultipart()

	if form.Name != "" {
		mp = mp.WithFormField("name", form.Name)
	}
	if form.Roaster != "" {
		mp = mp.WithFormField("roaster", form.Roaster)
	}
	if form.Region != "" {
		mp = mp.WithFormField("region", form.Region)
	}
	if form.Process != "" {
		mp = mp.WithFormField("process", form.Process)
	}
	if form.Description != "" {
		mp = mp.WithFormField("description", form.Description)
	}
	if form.OriginType != "" {
		mp = mp.WithFormField("origin_type", form.OriginType)
	}
	if form.RoastLevel != "" {
		mp = mp.WithFormField("roast_level", form.RoastLevel)
	}
	if form.Variety != "" {
		mp = mp.WithFormField("variety", form.Variety)
	}

	// Include numeric/bool fields when non-zero/true.
	// If your API needs explicit zero values, change this behavior.
	if form.Rating != 0 {
		mp = mp.WithFormField("rating", strconv.Itoa(form.Rating))
	}
	if form.Cost != 0 {
		mp = mp.WithFormField("cost", strconv.FormatFloat(form.Cost, 'f', -1, 64))
	}
	if form.Decaf {
		mp = mp.WithFormField("decaf", "true")
	}

	for _, note := range form.TastingNotes {
		mp = mp.WithFormField("tasting_notes", note)
	}

	if len(form.Img) > 0 {
		mp = mp.WithFileBytes("img", "test-image.jpg", form.Img)
	}

	for k, v := range extraFields {
		mp = mp.WithFormField(k, v)
	}

	return mp
}
