package api

import (
	"net/http"
	"testing"

	"github.com/Three-Ships/iterable-go/errors"
	"github.com/Three-Ships/iterable-go/logger"
	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"

	"github.com/stretchr/testify/assert"
)

func TestNewSubscriptionsApi(t *testing.T) {
	client := &http.Client{}
	api := NewSubscriptionsApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	assert.NotNil(t, api)
	assert.NotNil(t, api.api)
	assert.Equal(t, testApiKey, api.api.apiKey)
	assert.Equal(t, client, api.api.httpClient)
}

func TestSubscriptions_UnsubscribeChannelByUserID(t *testing.T) {
	testCases := []struct {
		name       string
		channelID  int64
		userID     string
		resBody    []byte
		resCode    int
		resErr     error
		expectURL  string
		expectRes  *types.PostResponse
		expectErr  bool
		resErrType string
	}{
		{
			name:      "successful response",
			channelID: 123,
			userID:    "user@example.com",
			resBody:   []byte(`{"msg":"User unsubscribed","code":"Success"}`),
			resCode:   http.StatusOK,
			expectURL: "https://api.iterable.com/api/subscriptions/messageChannel/123/byUserId/user@example.com",
			expectRes: &types.PostResponse{Message: "User unsubscribed", Code: "Success"},
		},
		{
			name:      "accepted response",
			channelID: 123,
			userID:    "user@example.com",
			resCode:   http.StatusAccepted,
			expectURL: "https://api.iterable.com/api/subscriptions/messageChannel/123/byUserId/user@example.com",
			expectRes: &types.PostResponse{},
		},
		{
			name:      "escaped user ID",
			channelID: 123,
			userID:    "user/name?query=value",
			resBody:   []byte(`{"msg":"User unsubscribed"}`),
			resCode:   http.StatusOK,
			expectURL: "https://api.iterable.com/api/subscriptions/messageChannel/123/byUserId/user%2Fname%3Fquery=value",
			expectRes: &types.PostResponse{Message: "User unsubscribed"},
		},
		{
			name:       "server error",
			channelID:  123,
			userID:     "user@example.com",
			resBody:    []byte(`{"message":"Internal Server Error"}`),
			resCode:    http.StatusInternalServerError,
			expectURL:  "https://api.iterable.com/api/subscriptions/messageChannel/123/byUserId/user@example.com",
			expectErr:  true,
			resErrType: errors.TYPE_HTTP_STATUS,
		},
		{
			name:       "network error",
			channelID:  123,
			userID:     "user@example.com",
			resErr:     assert.AnError,
			expectURL:  "https://api.iterable.com/api/subscriptions/messageChannel/123/byUserId/user@example.com",
			expectErr:  true,
			resErrType: errors.TYPE_IO,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := httpClient(tt.resBody, tt.resCode, tt.resErr)
			api := NewSubscriptionsApi(testApiKey, c, &logger.Noop{}, &rate.NoopLimiter{})

			res, err := api.UnsubscribeChannelByUserID(tt.channelID, tt.userID)
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
			assert.Equal(t, tt.expectURL, tr.Url())
			assert.Equal(t, http.MethodDelete, tr.Method())
			assert.Equal(t, testApiKey, tr.ApiKey())
		})
	}
}

func TestSubscriptions_UnsubscribeMessageTypeByUserID(t *testing.T) {
	c := httpClient([]byte(`{"msg":"User unsubscribed","code":"Success"}`), http.StatusOK, nil)
	api := NewSubscriptionsApi(testApiKey, c, &logger.Noop{}, &rate.NoopLimiter{})

	res, err := api.UnsubscribeMessageTypeByUserID(456, "user@example.com")

	assert.NoError(t, err)
	assert.Equal(t, &types.PostResponse{Message: "User unsubscribed", Code: "Success"}, res)
	tr, _ := c.Transport.(*testTransport)
	assert.Equal(t, "https://api.iterable.com/api/subscriptions/messageType/456/byUserId/user@example.com", tr.Url())
	assert.Equal(t, http.MethodDelete, tr.Method())
	assert.Equal(t, testApiKey, tr.ApiKey())
}
