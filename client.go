package iterable_go

import (
	"net/http"

	"github.com/Three-Ships/iterable-go/api"
)

type Client struct {
	httpClient *http.Client

	campaigns     *api.Campaigns
	catalog       *api.Catalog
	lists         *api.Lists
	channels      *api.Channels
	users         *api.Users
	events        *api.Events
	messageTypes  *api.MessageTypes
	subscriptions *api.Subscriptions
	templates     *api.Templates
}

func NewClient(apiKey string, opts ...ConfigOption) *Client {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := &http.Client{}
	httpClient.Transport = cfg.transport
	httpClient.Timeout = cfg.timeout

	return &Client{
		httpClient:    httpClient,
		campaigns:     api.NewCampaignsApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		catalog:       api.NewCatalogApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		lists:         api.NewListsApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		channels:      api.NewChannelsApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		users:         api.NewUsersApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		events:        api.NewEventsApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		messageTypes:  api.NewMessageTypesApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		subscriptions: api.NewSubscriptionsApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
		templates:     api.NewTemplatesApiWithOptions(apiKey, httpClient, cfg.logger, cfg.limiter, cfg.apiOpts...),
	}
}

func (c *Client) Campaigns() *api.Campaigns {
	return c.campaigns
}

func (c *Client) Catalog() *api.Catalog {
	return c.catalog
}

func (c *Client) Lists() *api.Lists {
	return c.lists
}

func (c *Client) Channels() *api.Channels {
	return c.channels
}

func (c *Client) Users() *api.Users {
	return c.users
}

func (c *Client) Events() *api.Events {
	return c.events
}

func (c *Client) MessageTypes() *api.MessageTypes {
	return c.messageTypes
}

func (c *Client) Subscriptions() *api.Subscriptions {
	return c.subscriptions
}

func (c *Client) Templates() *api.Templates {
	return c.templates
}
