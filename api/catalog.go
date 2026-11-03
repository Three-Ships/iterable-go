package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Three-Ships/iterable-go/logger"
	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"
)

var (
	PathCatalogs            = "catalogs"
	PathCatalogFieldMapping = "catalogs/{catalogName}/fieldMappings"
	PathCatalog             = "catalogs/{catalogName}"
	PathCatalogItem         = "catalogs/{catalogName}/items/{itemId}"
	PathCatalogItems        = "catalogs/{catalogName}/items"
)

// Catalog implements a set of /api/catalogs API methods,
// See: https://api.iterable.com/api/docs#catalogs_listCatalogs
type Catalog struct {
	api *apiClient
}

func NewCatalogApi(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter) *Catalog {
	return &Catalog{
		api: newApiClient(apiKey, httpClient, logger, limiter),
	}
}

func (c *Catalog) All() (*types.Catalogs, error) {
	var response types.Catalogs
	return toNilErr(&response, c.api.getJson(PathCatalogs, &response))
}

// ReplaceItem creates or replaces an item, deleting fields omitted from req.Value.
// Success means Iterable accepted the asynchronous write; visibility may lag.
func (c *Catalog) ReplaceItem(catalogName, itemID string, req types.CatalogItemReplaceRequest) (*types.PostResponse, error) {
	var response types.PostResponse
	path := strings.NewReplacer(
		"{catalogName}", url.PathEscape(catalogName),
		"{itemId}", url.PathEscape(itemID),
	).Replace(PathCatalogItem)
	return toNilErr(&response, c.api.sendJson(http.MethodPut, path, req, &response))
}

func (c *Catalog) FieldMapping(catalogName string) (*types.CatalogFieldMapping, error) {
	var response types.CatalogFieldMapping
	path := strings.Replace(PathCatalogFieldMapping, "{catalogName}", catalogName, 1)
	return toNilErr(&response, c.api.getJson(path, &response))
}
