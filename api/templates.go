package api

import (
	"net/http"

	"github.com/Three-Ships/iterable-go/logger"
	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"
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

// All retrieves all project templates across all available pages.
func (t *Templates) All() ([]types.Template, error) {
	templates := make([]types.Template, 0)

	err := paginate(PathTemplates+"?page=1&pageSize=1000&sort=id", func(path string) (string, error) {
		var res types.TemplatesResponse
		if err := t.api.getJson(path, &res); err != nil {
			return "", err
		}
		templates = append(templates, res.Templates...)
		return res.NextPageUrl, nil
	})
	if err != nil {
		return nil, err
	}
	return templates, nil
}
