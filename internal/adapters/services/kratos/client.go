package kratos

import (
	"github.com/lifenetwork-ai/iam-service/conf"
	kratos "github.com/ory/kratos-client-go"
)

// Client wraps the Kratos client configuration
type Client struct {
	publicAPI *kratos.APIClient
	adminAPI  *kratos.APIClient
}

// NewClient creates a new Kratos client
func NewClient(cfg *conf.KratosConfiguration) (*Client, error) {
	publicConfig := kratos.NewConfiguration()
	publicConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: cfg.PublicEndpoint,
		},
	}

	adminConfig := kratos.NewConfiguration()
	adminConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: cfg.AdminEndpoint,
		},
	}

	return &Client{
		publicAPI: kratos.NewAPIClient(publicConfig),
		adminAPI:  kratos.NewAPIClient(adminConfig),
	}, nil
}

// PublicAPI returns the public API client
func (c *Client) PublicAPI() *kratos.APIClient {
	return c.publicAPI
}

// AdminAPI returns the admin API client
func (c *Client) AdminAPI() *kratos.APIClient {
	return c.adminAPI
}
