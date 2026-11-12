package api

import (
	"net/http"

	"github.com/Three-Ships/iterable-go/logger"
	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"
)

var (
	PathChannels = "channels"
)

// Channels implements a set of /api/channels API methods,
// See: https://api.iterable.com/api/docs#channels_channels
type Channels struct {
	api *apiClient
}

func NewChannelsApi(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter) *Channels {
	return NewChannelsApiWithOptions(apiKey, httpClient, logger, limiter)
}

// NewChannelsApiWithOptions creates a channels API with the given options.
func NewChannelsApiWithOptions(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter, opts ...Option) *Channels {
	return &Channels{
		api: newApiClient(apiKey, httpClient, logger, limiter, opts...),
	}
}

func (c *Channels) Channels() ([]types.Channel, error) {
	var response types.ChannelsResponse
	return toNilErr(response.Channels, c.api.getJson(PathChannels, &response))
}
