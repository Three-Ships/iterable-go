package iterable_go

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Three-Ships/iterable-go/rate"
	"github.com/Three-Ships/iterable-go/types"
	"github.com/stretchr/testify/assert"
)

var (
	apiKey = "__API__KEY__"
)

func Test_newClient(t *testing.T) {
	c := NewClient(apiKey)
	assert.NotNil(t, c)
	assert.Equal(t, 10*time.Second, c.httpClient.Timeout)
	assert.NotNil(t, 10*time.Second, c.httpClient.Transport)
}

func Test_newClient_opts(t *testing.T) {
	tt := &fakeTransport{}
	c := NewClient(
		apiKey,
		WithTimeout(1*time.Second),
		WithTransport(tt),
		WithRateLimiter(&rate.NoopLimiter{}),
	)
	assert.Equal(t, 1*time.Second, c.httpClient.Timeout)
	assert.Equal(t, tt, c.httpClient.Transport)
}

func TestClientRegions(t *testing.T) {
	for _, region := range []struct {
		name string
		host string
		opts []ConfigOption
	}{
		{name: "US default", host: "api.iterable.com"},
		{name: "EU", host: "api.eu.iterable.com", opts: []ConfigOption{WithEU()}},
	} {
		t.Run(region.name, func(t *testing.T) {
			transport := &regionTransport{}
			client := NewClient(apiKey, append(region.opts, WithTransport(transport))...)
			calls := []struct {
				path string
				call func() error
			}{
				{"/api/campaigns", func() error { _, err := client.Campaigns().All(); return err }},
				{"/api/catalogs", func() error { _, err := client.Catalog().All(); return err }},
				{"/api/lists", func() error { _, err := client.Lists().All(); return err }},
				{"/api/channels", func() error { _, err := client.Channels().Channels(); return err }},
				{"/api/users/getFields", func() error { _, err := client.Users().GetAllFields(); return err }},
				{"/api/events/trackBulk", func() error { _, err := client.Events().TrackBulk(types.EventTrackBulkRequest{}); return err }},
				{"/api/messageTypes", func() error { _, err := client.MessageTypes().Get(); return err }},
				{"/api/subscriptions/messageChannel/1/byUserId/user", func() error { _, err := client.Subscriptions().UnsubscribeChannelByUserID(1, "user"); return err }},
				{"/api/templates", func() error { _, err := client.Templates().Get(); return err }},
			}
			for _, call := range calls {
				t.Run(call.path, func(t *testing.T) {
					transport.request = nil
					assert.NoError(t, call.call())
					if assert.NotNil(t, transport.request) {
						assert.Equal(t, "https", transport.request.URL.Scheme)
						assert.Equal(t, region.host, transport.request.URL.Host)
						assert.Equal(t, call.path, transport.request.URL.Path)
						assert.Equal(t, apiKey, transport.request.Header.Get("Api-Key"))
					}
				})
			}
		})
	}
}

type regionTransport struct{ request *http.Request }

func (t *regionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.request = req
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
}

func Test_newClient_init_all_apis(t *testing.T) {
	c := NewClient(apiKey)
	values := reflect.ValueOf(*c)
	types := reflect.TypeOf(*c)
	for i := range values.NumField() {
		field := values.Field(i)
		fieldName := types.Field(i).Name
		if field.IsNil() {
			assert.Fail(t, fmt.Sprintf("%s is not initialized", fieldName))
		}
	}
}

func Test_config_WithTransport(t *testing.T) {
	c := config{}
	WithTransport(&fakeTransport{})(&c)
	assert.NotNil(t, c.transport)
}

func Test_config_WithTimeout(t *testing.T) {
	c := config{}
	WithTimeout(2 * time.Second)(&c)
	assert.Equal(t, 2*time.Second, c.timeout)
}

func Test_config_WithRateLimiter(t *testing.T) {
	c := config{}
	WithRateLimiter(&rate.NoopLimiter{})(&c)
	assert.NotNil(t, c.limiter)
}

type fakeTransport struct {
}

func (f fakeTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, nil
}

var _ http.RoundTripper = &fakeTransport{}
