package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/block/iterable-go/errors"
	"github.com/block/iterable-go/logger"
	"github.com/block/iterable-go/rate"
)

const (
	baseUrl = "https://api.iterable.com/api"
)

type apiClient struct {
	apiKey     string
	httpClient *http.Client
	logger     logger.Logger
	limiter    rate.Limiter
}

func newApiClient(
	apiKey string,
	httpClient *http.Client,
	logger logger.Logger,
	limiter rate.Limiter,
) *apiClient {
	return &apiClient{
		apiKey:     apiKey,
		httpClient: httpClient,
		logger:     logger,
		limiter:    limiter,
	}
}

func (c *apiClient) getJson(path string, resData any) *errors.ApiError {
	return c.sendJson(http.MethodGet, path, nil, resData)
}

func (c *apiClient) postJson(path string, reqData, resData any) *errors.ApiError {
	return c.sendJson(http.MethodPost, path, reqData, resData)
}

func (c *apiClient) deleteJson(path string, reqData, resData any) *errors.ApiError {
	return c.sendJson(http.MethodDelete, path, reqData, resData)
}

func (c *apiClient) sendJson(
	httpMethod string,
	path string,
	reqData any,
	resData any,
) *errors.ApiError {
	body, err := c.send(
		httpMethod,
		path,
		reqData,
		"application/json",
		"application/json",
	)
	if err != nil {
		if len(err.Body) > 0 {
			code := iterableErr{}
			err2 := json.Unmarshal(err.Body, &code)
			if err2 == nil {
				err.IterableCode = code.Code
			}
			// Best effort to return some data
			_ = json.Unmarshal(body, resData)
		}
		return err
	}
	jsonErr := json.Unmarshal(body, resData)
	if jsonErr != nil {
		return &errors.ApiError{
			Stage:          errors.STAGE_AFTER_REQUEST,
			Type:           errors.TYPE_JSON_PARSE,
			SourceErr:      jsonErr,
			Body:           body,
			HttpStatusCode: http.StatusOK,
		}
	}
	return nil
}

func (c *apiClient) getText(path string) (string, *errors.ApiError) {
	return c.sendText(http.MethodGet, path, nil)
}

func (c *apiClient) sendText(
	httpMethod string,
	path string,
	reqData any,
) (string, *errors.ApiError) {
	body, err := c.send(
		httpMethod,
		path,
		reqData,
		"text/plain",
		"text/plain",
	)
	if body == nil {
		return "", err
	}
	return string(body), err
}

func (c *apiClient) send(
	httpMethod string,
	path string,
	reqData any,
	contentType string,
	accept string,
) ([]byte, *errors.ApiError) {
	endpoint := baseUrl + "/" + path

	var err error
	var req *http.Request

	if reqData != nil {
		data, jsonErr := json.Marshal(reqData)
		if jsonErr != nil {
			return nil, &errors.ApiError{
				Stage:     errors.STAGE_BEFORE_REQUEST,
				Type:      errors.TYPE_JSON_PARSE,
				SourceErr: jsonErr,
			}
		}
		req, err = http.NewRequest(
			httpMethod, endpoint, bytes.NewBuffer(data),
		)
	} else {
		req, err = http.NewRequest(
			httpMethod, endpoint, nil,
		)
	}

	if err != nil {
		return nil, &errors.ApiError{
			Stage:     errors.STAGE_BEFORE_REQUEST,
			Type:      errors.TYPE_REQUEST_PREP,
			SourceErr: err,
		}
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Api-Key", c.apiKey)
	req.Header.Set("Accept", accept)

	c.limiter.Limit(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &errors.ApiError{
			Stage:     errors.STAGE_REQUEST,
			Type:      errors.TYPE_IO,
			SourceErr: err,
		}
	}

	if res.StatusCode != http.StatusOK {
		var body []byte
		if res.Body != nil {
			body, _ = io.ReadAll(res.Body)
			defer func() { _ = res.Body.Close() }()
		}
		return body, &errors.ApiError{
			Stage:          errors.STAGE_AFTER_REQUEST,
			Type:           errors.TYPE_HTTP_STATUS,
			Body:           body,
			HttpStatusCode: res.StatusCode,
			SourceErr:      err,
		}
	}

	body, err := io.ReadAll(res.Body)
	defer func() { _ = res.Body.Close() }()
	if err != nil {
		return body, &errors.ApiError{
			Stage:          errors.STAGE_AFTER_REQUEST,
			Type:           errors.TYPE_IO,
			Body:           body,
			HttpStatusCode: res.StatusCode,
			SourceErr:      err,
		}
	}

	return body, nil
}

// toNilErr converts a *errors.ApiError type to be a true nil interface.
// Internally, a Go interface has a Type and Value.
// An interface value is nil only if the V and T are both unset.
// See: https://go.dev/doc/faq#nil_error
func toNilErr[T any](r T, e *errors.ApiError) (T, error) {
	if e != nil {
		return r, e
	}
	return r, nil
}

// notImplemented is used as a temporary solution for methods
// that will be or are being implemented.
//
// Usage:
//
//	func (c *Catalog) Delete(catalogName string) error {
//		return notImplemented(http.MethodDelete, PathCatalog)
//	}
func notImplemented(httpMethod string, endpoint string) error {
	return &errors.ApiError{
		Stage: errors.STAGE_BEFORE_REQUEST,
		Type:  errors.TYPE_NOT_IMPLEMENTED,
		SourceErr: fmt.Errorf(
			"%s %s is not implemented", httpMethod, endpoint,
		),
	}
}

func validatePaginationPath(raw string, expectedEndpoint string) (string, string, error) {
	if raw == "" {
		return "", "", fmt.Errorf("empty url")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("parse url: %w", err)
	}

	if parsed.Scheme != "" && !strings.EqualFold(parsed.Scheme, "https") {
		return "", "", fmt.Errorf("unsupported url scheme %q", parsed.Scheme)
	}
	if parsed.Host != "" {
		if !strings.EqualFold(parsed.Hostname(), "api.iterable.com") {
			return "", "", fmt.Errorf("untrusted url host %q", parsed.Host)
		}
		if parsed.Port() != "" && parsed.Port() != "443" {
			return "", "", fmt.Errorf("unsupported url port %q", parsed.Port())
		}
	}
	if parsed.User != nil {
		return "", "", fmt.Errorf("url userinfo is not allowed")
	}

	cleanPath := path.Clean(parsed.Path)
	trimmed := strings.TrimPrefix(cleanPath, "/api/")
	trimmed = strings.TrimPrefix(trimmed, "api/")
	trimmed = strings.TrimPrefix(trimmed, "/")

	if trimmed != expectedEndpoint {
		return "", "", fmt.Errorf("unexpected endpoint path %q (expected %q)", trimmed, expectedEndpoint)
	}

	reqPath := trimmed
	normKey := trimmed
	if parsed.RawQuery != "" {
		reqPath += "?" + parsed.RawQuery
		normKey += "?" + parsed.Query().Encode()
	}
	return reqPath, normKey, nil
}

const (
	maxPaginationPages = 1000
)

func paginate[T any, R any](
	client *apiClient,
	initialPath string,
	expectedEndpoint string,
	extract func(*R) ([]T, string),
) ([]T, error) {
	all := make([]T, 0)
	initialReqPath, initialKey, err := validatePaginationPath(initialPath, expectedEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid initial pagination path: %w", err)
	}
	seenURLs := map[string]struct{}{initialKey: {}}
	reqPath := initialReqPath

	for page := 0; page < maxPaginationPages; page++ {
		var res R
		if err := client.getJson(reqPath, &res); err != nil {
			return nil, err
		}
		items, nextURL := extract(&res)
		all = append(all, items...)
		if nextURL == "" {
			return all, nil
		}
		nextReqPath, nextKey, err := validatePaginationPath(nextURL, expectedEndpoint)
		if err != nil {
			return nil, fmt.Errorf("invalid next page url: %w", err)
		}
		if _, seen := seenURLs[nextKey]; seen {
			return nil, fmt.Errorf("repeated next page url for %s", expectedEndpoint)
		}
		seenURLs[nextKey] = struct{}{}
		reqPath = nextReqPath
	}
	return nil, fmt.Errorf("pagination exceeded maximum limit of %d pages", maxPaginationPages)
}

type iterableErr struct {
	Code string `json:"code"`
}
