package api

import (
	"net/http"

	"github.com/block/iterable-go/logger"
	"github.com/block/iterable-go/rate"
	"github.com/block/iterable-go/types"
)

var (
	PathTemplates = "templates"
)

// Templates implements a set of /api/templates API methods,
// See: https://api.iterable.com/api/docs#templates_templates
type Templates struct {
	api *apiClient
}

func NewTemplatesApi(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter) *Templates {
	return &Templates{
		api: newApiClient(apiKey, httpClient, logger, limiter),
	}
}

func (t *Templates) Get() (*types.TemplatesResponse, error) {
	var res types.TemplatesResponse
	return toNilErr(&res, t.api.getJson(PathTemplates, &res))
}

func (t *Templates) All() ([]types.Template, error) {
	allTemplates := make([]types.Template, 0)
	path := PathTemplates + "?page=1&pageSize=1000&sort=id"
	seenUrls := make(map[string]struct{})

	for {
		var res types.TemplatesResponse
		if err := t.api.getJson(path, &res); err != nil {
			return nil, err
		}
		allTemplates = append(allTemplates, res.Templates...)
		if res.NextPageUrl == "" {
			break
		}
		path = sanitizeApiPath(res.NextPageUrl)
		if _, seen := seenUrls[path]; seen {
			break
		}
		seenUrls[path] = struct{}{}
	}
	return allTemplates, nil
}
