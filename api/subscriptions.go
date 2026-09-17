package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/Three-Ships/iterable-go/logger"
	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"
)

const (
	pathSubscriptionsMessageChannel = "subscriptions/messageChannel"
	pathSubscriptionsMessageType    = "subscriptions/messageType"
)

// Subscriptions implements a set of /api/subscriptions API methods.
type Subscriptions struct {
	api *apiClient
}

func NewSubscriptionsApi(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter) *Subscriptions {
	return &Subscriptions{
		api: newApiClient(apiKey, httpClient, logger, limiter),
	}
}

// UnsubscribeChannelByUserID prevents delivery from every message type in the channel.
func (s *Subscriptions) UnsubscribeChannelByUserID(channelID int64, userID string) (*types.PostResponse, error) {
	return s.unsubscribeByUserID(pathSubscriptionsMessageChannel, channelID, userID)
}

// UnsubscribeMessageTypeByUserID unsubscribes a user from one message type.
func (s *Subscriptions) UnsubscribeMessageTypeByUserID(messageTypeID int64, userID string) (*types.PostResponse, error) {
	return s.unsubscribeByUserID(pathSubscriptionsMessageType, messageTypeID, userID)
}

func (s *Subscriptions) unsubscribeByUserID(pathPrefix string, subscriptionGroupID int64, userID string) (*types.PostResponse, error) {
	path := pathPrefix + "/" + strconv.FormatInt(subscriptionGroupID, 10) + "/byUserId/" + url.PathEscape(userID)
	var res types.PostResponse
	return toNilErr(&res, s.api.deleteJson(path, nil, &res))
}
