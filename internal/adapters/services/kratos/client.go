package kratos

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/conf"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	kratos "github.com/ory/kratos-client-go"
)

// Client wraps the Kratos client configuration
type Client struct {
	tenantRepo domainrepo.TenantRepository
	clientMap  sync.Map // map[uuid.UUID]*tenantClient
	config     *conf.KratosConfiguration
}

type tenantClient struct {
	publicAPI *kratos.APIClient
	adminAPI  *kratos.APIClient
}

// NewClient creates a new Kratos client
func NewClient(cfg *conf.KratosConfiguration, tenantRepo domainrepo.TenantRepository) *Client {
	// Initialize client map
	client := &Client{}
	client.tenantRepo = tenantRepo
	client.config = cfg
	tenants, err := tenantRepo.List()
	if err != nil {
		logger.GetLogger().Errorf("failed to get tenants: %v", err)
	}
	for _, tenant := range tenants {
		err := client.InitializeTenantClient(context.Background(), tenant.ID)
		if err != nil {
			logger.GetLogger().Errorf("failed to initialize tenant client: %v", err)
		}
	}
	return client
}

// InitializeTenantClient initializes a client for a specific tenant
func (c *Client) InitializeTenantClient(ctx context.Context, tenantID uuid.UUID) error {
	// Get tenant info from repository
	tenant, err := c.tenantRepo.GetByID(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	// Create new client for tenant
	publicConfig := kratos.NewConfiguration()
	publicConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: tenant.PublicURL,
		},
	}

	adminConfig := kratos.NewConfiguration()
	adminConfig.Servers = []kratos.ServerConfiguration{
		{
			URL: tenant.AdminURL,
		},
	}

	client := &tenantClient{
		publicAPI: kratos.NewAPIClient(publicConfig),
		adminAPI:  kratos.NewAPIClient(adminConfig),
	}
	// Store in map
	c.clientMap.Store(tenantID, client)
	return nil
}

// GetTenantClient gets a tenant-specific client
func (c *Client) GetTenantClient(tenantID uuid.UUID) (*tenantClient, error) {
	// Check if client exists in map
	if client, ok := c.clientMap.Load(tenantID); ok {
		return client.(*tenantClient), nil
	}

	// Initialize client if it doesn't exist
	if err := c.InitializeTenantClient(context.Background(), tenantID); err != nil {
		return nil, err
	}

	// Get the newly initialized client
	if client, ok := c.clientMap.Load(tenantID); ok {
		return client.(*tenantClient), nil
	}

	return nil, fmt.Errorf("failed to get tenant client: %s", tenantID)
}

// PublicAPI returns the public API client for a specific tenant
func (c *Client) PublicAPI(tenantID uuid.UUID) (*kratos.APIClient, error) {
	client, err := c.GetTenantClient(tenantID)
	if err != nil {
		return nil, err
	}
	return client.publicAPI, nil
}

// AdminAPI returns the admin API client for a specific tenant
func (c *Client) AdminAPI(tenantID uuid.UUID) (*kratos.APIClient, error) {
	client, err := c.GetTenantClient(tenantID)
	if err != nil {
		return nil, err
	}
	return client.adminAPI, nil
}
