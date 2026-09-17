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

// All retrieves all project templates across all available pages (up to 1,000 pages).
func (t *Templates) All() ([]types.Template, error) {
	return paginate(
		t.api,
		PathTemplates+"?page=1&pageSize=1000&sort=id",
		PathTemplates,
		func(res *types.TemplatesResponse) ([]types.Template, string) {
			return res.Templates, res.NextPageUrl
		},
	)
}
