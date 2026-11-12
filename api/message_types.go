package api

import (
	"net/http"

	"github.com/Three-Ships/iterable-go/logger"
	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"
)

const pathMessageTypes = "messageTypes"

// MessageTypes implements a set of /api/messageTypes API methods,
// See: https://api.iterable.com/api/docs#messageTypes_messageTypes
type MessageTypes struct {
	api *apiClient
}

func NewMessageTypesApi(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter) *MessageTypes {
	return NewMessageTypesApiWithOptions(apiKey, httpClient, logger, limiter)
}

// NewMessageTypesApiWithOptions creates a message types API with the given options.
func NewMessageTypesApiWithOptions(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter, opts ...Option) *MessageTypes {
	return &MessageTypes{
		api: newApiClient(apiKey, httpClient, logger, limiter, opts...),
	}
}

func (m *MessageTypes) Get() (*types.MessageTypeResponse, error) {
	var res types.MessageTypeResponse
	return toNilErr(&res, m.api.getJson(
		pathMessageTypes, &res,
	))
}
