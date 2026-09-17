package api

import (
	"bytes"
	"net/http"
	"sync"
	"testing"

	"github.com/block/iterable-go/errors"
	"github.com/block/iterable-go/logger"
	"github.com/block/iterable-go/rate"
	"github.com/block/iterable-go/types"

	"github.com/stretchr/testify/assert"
)

func TestNewTemplatesApi(t *testing.T) {
	client := &http.Client{}
	api := NewTemplatesApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	assert.NotNil(t, api)
	assert.NotNil(t, api.api)
	assert.Equal(t, testApiKey, api.api.apiKey)
	assert.Equal(t, client, api.api.httpClient)
}

func TestTemplates_Get(t *testing.T) {
	testCases := []struct {
		name       string
		resBody    []byte
		resCode    int
		resErr     error
		expectUrl  string
		expectRes  *types.TemplatesResponse
		expectErr  bool
		resErrType string
	}{
		{
			name: "successful response",
			resBody: []byte(`{
				"templates": [{
					"templateId": 1,
					"name": "Welcome",
					"createdAt": "2026-09-01 12:00:00 +00:00",
					"updatedAt": "2026-09-02 12:00:00 +00:00",
					"creatorUserId": "creator@example.com",
					"messageTypeId": 2
				}]
			}`),
			resCode:   200,
			expectUrl: "https://api.iterable.com/api/templates",
			expectRes: &types.TemplatesResponse{
				Templates: []types.Template{{
					TemplateId: 1, Name: "Welcome", CreatedAt: "2026-09-01 12:00:00 +00:00",
					UpdatedAt: "2026-09-02 12:00:00 +00:00", CreatorUserId: "creator@example.com", MessageTypeId: 2,
				}},
			},
		},
		{
			name:      "empty templates",
			resBody:   []byte(`{"templates": []}`),
			resCode:   200,
			expectUrl: "https://api.iterable.com/api/templates",
			expectRes: &types.TemplatesResponse{Templates: []types.Template{}},
		},
		{
			name:       "malformed json response",
			resBody:    []byte(`{"templates": [{]}`),
			resCode:    200,
			expectUrl:  "https://api.iterable.com/api/templates",
			expectErr:  true,
			resErrType: errors.TYPE_JSON_PARSE,
		},
		{
			name:       "server error",
			resBody:    []byte(`{"message": "Internal Server Error"}`),
			resCode:    500,
			expectUrl:  "https://api.iterable.com/api/templates",
			expectErr:  true,
			resErrType: errors.TYPE_HTTP_STATUS,
		},
		{
			name:       "network error",
			resErr:     assert.AnError,
			resCode:    0,
			expectUrl:  "https://api.iterable.com/api/templates",
			expectErr:  true,
			resErrType: errors.TYPE_IO,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := httpClient(tt.resBody, tt.resCode, tt.resErr)
			api := NewTemplatesApi(testApiKey, c, &logger.Noop{}, &rate.NoopLimiter{})

			res, err := api.Get()
			if tt.expectErr {
				assert.Error(t, err)
				apiError := err.(*errors.ApiError)
				assert.Equal(t, tt.resCode, apiError.HttpStatusCode)
				assert.Equal(t, tt.resErrType, apiError.Type)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRes, res)
			}

			tr, _ := c.Transport.(*testTransport)
			assert.Equal(t, tt.expectUrl, tr.Url())
			assert.Equal(t, http.MethodGet, tr.Method())
			assert.Equal(t, testApiKey, tr.ApiKey())
		})
	}
}

func TestTemplates_All(t *testing.T) {
	testCases := []struct {
		name       string
		resBody    []byte
		resCode    int
		resErr     error
		expectUrl  string
		expectRes  []types.Template
		expectErr  bool
		resErrType string
	}{
		{
			name:      "successful response",
			resBody:   []byte(`{"templates": [{"templateId": 1, "name": "Welcome"}]}`),
			resCode:   200,
			expectUrl: "https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id",
			expectRes: []types.Template{{TemplateId: 1, Name: "Welcome"}},
		},
		{
			name:      "empty response",
			resBody:   []byte(`{"templates": []}`),
			resCode:   200,
			expectUrl: "https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id",
			expectRes: []types.Template{},
		},
		{
			name:       "malformed json",
			resBody:    []byte(`{"templates": [{"templateId":`),
			resCode:    200,
			expectUrl:  "https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id",
			expectErr:  true,
			resErrType: errors.TYPE_JSON_PARSE,
		},
		{
			name:       "server error",
			resBody:    []byte(`{"message": "Internal Server Error"}`),
			resCode:    500,
			expectUrl:  "https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id",
			expectErr:  true,
			resErrType: errors.TYPE_HTTP_STATUS,
		},
		{
			name:       "network error",
			resErr:     assert.AnError,
			resCode:    0,
			expectUrl:  "https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id",
			expectErr:  true,
			resErrType: errors.TYPE_IO,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := httpClient(tt.resBody, tt.resCode, tt.resErr)
			api := NewTemplatesApi(testApiKey, c, &logger.Noop{}, &rate.NoopLimiter{})

			templates, err := api.All()
			if tt.expectErr {
				assert.Error(t, err)
				apiError := err.(*errors.ApiError)
				assert.Equal(t, tt.resCode, apiError.HttpStatusCode)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRes, templates)
			}

			tr, _ := c.Transport.(*testTransport)
			assert.Equal(t, tt.expectUrl, tr.Url())
			assert.Equal(t, http.MethodGet, tr.Method())
			assert.Equal(t, testApiKey, tr.ApiKey())
		})
	}
}

func TestTemplates_All_Paginates(t *testing.T) {
	t.Parallel()

	transport := &templatePagesTransport{
		bodies: [][]byte{
			[]byte(`{"templates":[{"templateId":1,"name":"Template 1"}],"nextPageUrl":"/api/templates?page=2&pageSize=1000&sort=id"}`),
			[]byte(`{"templates":[{"templateId":2,"name":"Template 2"}]}`),
		},
	}
	client := &http.Client{Transport: transport}
	api := NewTemplatesApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	templates, err := api.All()
	assert.NoError(t, err)
	assert.Equal(t, []types.Template{
		{TemplateId: 1, Name: "Template 1"},
		{TemplateId: 2, Name: "Template 2"},
	}, templates)
	assert.Equal(t, []string{
		"https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id",
		"https://api.iterable.com/api/templates?page=2&pageSize=1000&sort=id",
	}, transport.urls)
}

type templatePagesTransport struct {
	mu     sync.Mutex
	bodies [][]byte
	urls   []string
}

func (t *templatePagesTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.urls = append(t.urls, request.URL.String())
	body := t.bodies[0]
	t.bodies = t.bodies[1:]
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{Reader: bytes.NewReader(body)},
	}, nil
}
