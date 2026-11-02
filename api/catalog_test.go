package api

import (
	"io"
	"net/http"
	"testing"

	"github.com/Three-Ships/iterable-go/errors"
	"github.com/Three-Ships/iterable-go/logger"
	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCatalogApi(t *testing.T) {
	client := &http.Client{}
	api := NewCatalogApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	assert.NotNil(t, api)
	assert.NotNil(t, api.api)
	assert.Equal(t, testApiKey, api.api.apiKey)
	assert.Equal(t, client, api.api.httpClient)
}

func TestCatalog_ReplaceItem(t *testing.T) {
	testCases := []struct {
		name       string
		resBody    string
		resCode    int
		resErr     error
		want       types.PostResponse
		wantErr    string
		wantStatus int
	}{
		{
			name:    "accepted asynchronously",
			resBody: `{"code":"Success","msg":"","params":{"catalogName":"weather","itemId":"edition"}}`,
			resCode: http.StatusAccepted,
			want:    types.PostResponse{Code: "Success", Params: map[string]interface{}{"catalogName": "weather", "itemId": "edition"}},
		},
		{
			name:    "successful response",
			resBody: `{"code":"Success","msg":"updated"}`,
			resCode: http.StatusOK,
			want:    types.PostResponse{Code: "Success", Message: "updated"},
		},
		{name: "empty accepted response", resCode: http.StatusAccepted},
		{name: "malformed response", resCode: http.StatusOK, resBody: `{`, wantErr: errors.TYPE_JSON_PARSE, wantStatus: http.StatusOK},
		{name: "invalid item", resCode: http.StatusBadRequest, resBody: `{"code":"BadParams","msg":"Invalid item"}`, wantErr: errors.TYPE_HTTP_STATUS, wantStatus: http.StatusBadRequest},
		{name: "unauthorized", resCode: http.StatusUnauthorized, wantErr: errors.TYPE_HTTP_STATUS, wantStatus: http.StatusUnauthorized},
		{name: "missing catalog", resCode: http.StatusNotFound, wantErr: errors.TYPE_HTTP_STATUS, wantStatus: http.StatusNotFound},
		{name: "rate limited", resCode: http.StatusTooManyRequests, resBody: "Too Many Requests", wantErr: errors.TYPE_HTTP_STATUS, wantStatus: http.StatusTooManyRequests},
		{name: "server failure", resCode: http.StatusServiceUnavailable, wantErr: errors.TYPE_HTTP_STATUS, wantStatus: http.StatusServiceUnavailable},
		{name: "network failure", resErr: assert.AnError, wantErr: errors.TYPE_IO},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := httpClient([]byte(tt.resBody), tt.resCode, tt.resErr)
			api := NewCatalogApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})
			req := types.CatalogItemReplaceRequest{
				Value: map[string]any{
					"days": []map[string]any{{
						"date": "2026-10-30",
						"segments": []map[string]any{{
							"temperature_f":        0,
							"precipitation_chance": nil,
						}},
					}},
				},
			}

			res, err := api.ReplaceItem("brand-weather", "edition-key", req)
			if tt.wantErr != "" {
				var apiErr *errors.ApiError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, tt.wantErr, apiErr.Type)
				assert.Equal(t, tt.wantStatus, apiErr.HttpStatusCode)
				if tt.resErr != nil {
					assert.ErrorIs(t, apiErr.SourceErr, tt.resErr)
				}
				if tt.resCode == http.StatusBadRequest {
					assert.Equal(t, "BadParams", apiErr.IterableCode)
					assert.Equal(t, "Invalid item", apiErr.IterableMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, &tt.want, res)
			}

			tr := client.Transport.(*testTransport)
			assert.Equal(t, "https://api.iterable.com/api/catalogs/brand-weather/items/edition-key", tr.Url())
			assert.Equal(t, http.MethodPut, tr.Method())
			assert.Equal(t, testApiKey, tr.ApiKey())
			assert.Equal(t, "application/json", tr.req.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", tr.req.Header.Get("Accept"))
			body, readErr := io.ReadAll(tr.req.Body)
			require.NoError(t, readErr)
			assert.JSONEq(t, `{"value":{"days":[{"date":"2026-10-30","segments":[{"temperature_f":0,"precipitation_chance":null}]}]}}`, string(body))
			if tt.resErr == nil {
				assert.True(t, tr.res.Body.(*testReader).isClosed)
			}
		})
	}
}

func TestCatalog_ReplaceItemValue(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
	}{
		{
			name:  "map",
			value: map[string]any{"name": "Example", "count": 0},
		},
		{
			name: "struct",
			value: struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}{Name: "Example"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := httpClient(nil, http.StatusAccepted, nil)
			api := NewCatalogApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

			_, err := api.ReplaceItem("catalog", "item", types.CatalogItemReplaceRequest{Value: tt.value})
			require.NoError(t, err)

			body, err := io.ReadAll(client.Transport.(*testTransport).req.Body)
			require.NoError(t, err)
			assert.JSONEq(t, `{"value":{"name":"Example","count":0}}`, string(body))
		})
	}
}

func TestCatalog_ReplaceItemUnencodableValue(t *testing.T) {
	client := httpClient(nil, http.StatusAccepted, nil)
	api := NewCatalogApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	_, err := api.ReplaceItem("weather", "edition", types.CatalogItemReplaceRequest{
		Value: map[string]any{"invalid": make(chan int)},
	})

	var apiErr *errors.ApiError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, errors.STAGE_BEFORE_REQUEST, apiErr.Stage)
	assert.Equal(t, errors.TYPE_JSON_PARSE, apiErr.Type)
	assert.Nil(t, client.Transport.(*testTransport).req)
}

func TestCatalog_ReplaceItemEscapesPathParameters(t *testing.T) {
	client := httpClient(nil, http.StatusAccepted, nil)
	api := NewCatalogApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	_, err := api.ReplaceItem("catalog/other", "item?query=value#fragment", types.CatalogItemReplaceRequest{
		Value: map[string]any{},
	})

	require.NoError(t, err)
	tr := client.Transport.(*testTransport)
	assert.Equal(t, "https://api.iterable.com/api/catalogs/catalog%2Fother/items/item%3Fquery=value%23fragment", tr.Url())
	assert.Empty(t, tr.req.URL.RawQuery)
	assert.Empty(t, tr.req.URL.Fragment)
}

func TestCatalog_All(t *testing.T) {
	testCases := []struct {
		name       string
		resBody    []byte
		resCode    int
		resErr     error
		expectUrl  string
		expectRes  *types.Catalogs
		expectErr  bool
		resErrType string
	}{
		{
			name: "successful response",
			resBody: []byte(`{
				"catalogNames": ["test-catalog"], "totalCatalogsCount": 100
			}`),
			resCode:   200,
			expectUrl: "https://api.iterable.com/api/catalogs",
			expectRes: &types.Catalogs{
				CatalogNames:       []string{"test-catalog"},
				TotalCatalogsCount: 100,
			},
		},
		{
			name:      "empty response",
			resBody:   []byte(`{"catalogNames":[]}`),
			resCode:   200,
			expectUrl: "https://api.iterable.com/api/catalogs",
			expectRes: &types.Catalogs{
				CatalogNames: []string{},
			},
		},
		{
			name:       "malformed json",
			resBody:    []byte(`{"catalogNames":`),
			resCode:    200,
			expectUrl:  "https://api.iterable.com/api/catalogs",
			expectErr:  true,
			resErrType: errors.TYPE_JSON_PARSE,
		},
		{
			name:       "server error",
			resBody:    []byte(`{"message": "Internal Server Error"}`),
			resCode:    500,
			expectUrl:  "https://api.iterable.com/api/catalogs",
			expectErr:  true,
			resErrType: errors.TYPE_HTTP_STATUS,
		},
		{
			name:       "server error",
			resBody:    []byte(`Too Many Requests`),
			resCode:    429,
			expectUrl:  "https://api.iterable.com/api/catalogs",
			expectErr:  true,
			resErrType: errors.TYPE_HTTP_STATUS,
		},
		{
			name:       "network error",
			resErr:     assert.AnError,
			resCode:    0,
			expectUrl:  "https://api.iterable.com/api/catalogs",
			expectErr:  true,
			resErrType: errors.TYPE_IO,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := httpClient(tt.resBody, tt.resCode, tt.resErr)
			api := NewCatalogApi(testApiKey, c, &logger.Noop{}, &rate.NoopLimiter{})

			catalogs, err := api.All()
			if tt.expectErr {
				assert.Error(t, err)
				apiError := err.(*errors.ApiError)
				assert.Equal(t, tt.resCode, apiError.HttpStatusCode)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRes, catalogs)
			}

			tr, _ := c.Transport.(*testTransport)
			assert.Equal(t, tt.expectUrl, tr.Url())
			assert.Equal(t, http.MethodGet, tr.Method())
			assert.Equal(t, testApiKey, tr.ApiKey())
		})
	}
}

func TestCatalog_FieldMapping(t *testing.T) {
	testCases := []struct {
		name        string
		catalogName string
		resBody     []byte
		resCode     int
		resErr      error
		expectUrl   string
		expectRes   *types.CatalogFieldMapping
		expectErr   bool
		resErrType  string
	}{
		{
			name:        "successful response",
			catalogName: "test-catalog",
			resBody: []byte(`{
				"definedMappings": {"field1": "a"},
				"undefinedFields": ["field2", "field3"]
			}`),
			resCode:   200,
			expectUrl: "https://api.iterable.com/api/catalogs/test-catalog/fieldMappings",
			expectRes: &types.CatalogFieldMapping{
				DefinedMappings: map[string]any{
					"field1": "a", // keep it at 1 to avoid sorting issues
				},
				UndefinedFields: []string{
					"field2",
					"field3",
				},
			},
		},
		{
			name:        "empty response",
			catalogName: "empty-catalog",
			resBody:     []byte(`{"definedMappings":{}}`),
			resCode:     200,
			expectUrl:   "https://api.iterable.com/api/catalogs/empty-catalog/fieldMappings",
			expectRes: &types.CatalogFieldMapping{
				DefinedMappings: map[string]interface{}{},
			},
		},
		{
			name:        "malformed json",
			catalogName: "bad-catalog",
			resBody:     []byte(`{"mappings": [{]}`),
			resCode:     200,
			expectUrl:   "https://api.iterable.com/api/catalogs/bad-catalog/fieldMappings",
			expectErr:   true,
			resErrType:  errors.TYPE_JSON_PARSE,
		},
		{
			name:        "server error",
			catalogName: "error-catalog",
			resBody:     []byte(`{"message": "Internal Server Error"}`),
			resCode:     500,
			expectUrl:   "https://api.iterable.com/api/catalogs/error-catalog/fieldMappings",
			expectErr:   true,
			resErrType:  errors.TYPE_HTTP_STATUS,
		},
		{
			name:        "429",
			catalogName: "error-catalog",
			resBody:     []byte(`Too Many Requests`),
			resCode:     429,
			expectUrl:   "https://api.iterable.com/api/catalogs/error-catalog/fieldMappings",
			expectErr:   true,
			resErrType:  errors.TYPE_HTTP_STATUS,
		},
		{
			name:        "network-error",
			catalogName: "error-catalog",
			resErr:      assert.AnError,
			resCode:     0,
			expectUrl:   "https://api.iterable.com/api/catalogs/error-catalog/fieldMappings",
			expectErr:   true,
			resErrType:  errors.TYPE_IO,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := httpClient(tt.resBody, tt.resCode, tt.resErr)
			api := NewCatalogApi(testApiKey, c, &logger.Noop{}, &rate.NoopLimiter{})

			mapping, err := api.FieldMapping(tt.catalogName)
			if tt.expectErr {
				assert.Error(t, err)
				apiError := err.(*errors.ApiError)
				assert.Equal(t, tt.resCode, apiError.HttpStatusCode)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRes, mapping)
			}

			tr, _ := c.Transport.(*testTransport)
			assert.Equal(t, tt.expectUrl, tr.Url())
			assert.Equal(t, http.MethodGet, tr.Method())
			assert.Equal(t, testApiKey, tr.ApiKey())
		})
	}
}
