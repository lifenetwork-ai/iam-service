package keto

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	domainservice "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
	ucasetypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	keto "github.com/ory/keto-client-go"
)

// Client wraps the Keto client configuration
type Client struct {
	tenantRepo domainrepo.TenantRepository
	client     *keto.APIClient
	config     *conf.KetoConfiguration
}

var _ domainservice.KetoService = (*Client)(nil)

// NewClient creates a new Keto client
func NewKetoService(tenantRepo domainrepo.TenantRepository) domainservice.KetoService {
	cfg := conf.GetKetoConfig()
	ketoCfg := keto.NewConfiguration()

	ketoCfg.OperationServers = map[string]keto.ServerConfigurations{
		// Write Operations
		constants.OperationCreateRelationship: {
			{
				URL:         cfg.DefaultWriteURL,
				Description: constants.KetoWriteApiDescription,
			},
		},
		constants.OperationDeleteRelationships: {
			{
				URL:         cfg.DefaultWriteURL,
				Description: constants.KetoWriteApiDescription,
			},
		},
		constants.OperationPatchRelationships: {
			{
				URL:         cfg.DefaultWriteURL,
				Description: constants.KetoWriteApiDescription,
			},
		},

		// Read Operations
		constants.OperationGetRelationships: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationListRelationshipNamespaces: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},

		// Permission Check Operations
		constants.OperationCheckPermission: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationCheckPermissionOrError: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationPostCheckPermission: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationPostCheckPermissionOrError: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationExpandPermissions: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationCheckOplSyntax: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},

		// Metadata/Health
		constants.OperationGetVersion: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationIsAlive: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
		constants.OperationIsReady: {
			{
				URL:         cfg.DefaultReadURL,
				Description: constants.KetoReadApiDescription,
			},
		},
	}

	client := keto.NewAPIClient(ketoCfg)

	return &Client{
		tenantRepo: tenantRepo,
		client:     client,
		config:     cfg,
	}
}

// CheckPermission checks if a subject has permission to perform an action on an object
func (c *Client) CheckPermission(ctx context.Context, request ucasetypes.CheckPermissionRequest) (bool, *domainerrors.DomainError) {
	req := c.client.PermissionApi.PostCheckPermission(ctx).PostCheckPermissionBody(c.toKetoCheckPermissionBody(request))
	ketoResp, _, err := req.Execute()
	if err != nil {
		logger.GetLogger().Errorf("failed to check permission: %v", err)
		return false, domainerrors.NewInternalError(
			"MSG_FAILED_TO_CHECK_PERMISSION",
			"Failed to check permission",
		)
	}

	return ketoResp.GetAllowed(), nil
}

// CreateRelationTuple creates a relation tuple
// Note: The dto should be validated before calling this function
func (c *Client) CreateRelationTuple(ctx context.Context, request ucasetypes.CreateRelationTupleRequest) *domainerrors.DomainError {
	logger.GetLogger().Debugf("Creating relation tuple for namespace: %s, object: %s, relation: %s, subject_set: %v",
		request.Namespace, request.Object, request.Relation, request.TenantRelation)

	req := c.client.RelationshipApi.CreateRelationship(ctx).CreateRelationshipBody(c.toKetoCreateRelationshipBody(request))

	// Log the request details before execution
	logger.GetLogger().Debugf("Sending request to Keto Write API URL: %s", c.config.DefaultWriteURL)
	fmt.Println("request", c.toKetoCreateRelationshipBody(request))
	fmt.Println("url", c.config.DefaultWriteURL)
	logger.GetLogger().Debugf("Request body: %+v", c.toKetoCreateRelationshipBody(request))

	_, httpResp, err := req.Execute()
	fmt.Println("httpResp", httpResp)
	if httpResp != nil {
		logger.GetLogger().Debugf("Response from Keto: Status: %d, Headers: %v", httpResp.StatusCode, httpResp.Header)
	}
	if err != nil {
		logger.GetLogger().Errorf("failed to create relation tuple: %v", err)
		return domainerrors.NewInternalError(
			"MSG_FAILED_TO_CREATE_RELATION_TUPLE",
			"Failed to create relation tuple",
		)
	}

	// Keto returns 201 Created on success
	if httpResp.StatusCode != http.StatusCreated {
		logger.GetLogger().Errorf("failed to create relation tuple: unexpected status code %d", httpResp.StatusCode)
		return domainerrors.NewInternalError(
			"MSG_FAILED_TO_CREATE_RELATION_TUPLE",
			"Failed to create relation tuple",
		)
	}

	logger.GetLogger().Debugf("Successfully created relation tuple for %s:%s#%s",
		request.Namespace, request.Object, request.Relation)
	return nil
}

func (c *Client) toKetoCheckPermissionBody(req ucasetypes.CheckPermissionRequest) keto.PostCheckPermissionBody {
	return keto.PostCheckPermissionBody{
		Namespace: &req.Namespace,
		Relation:  &req.Relation,
		Object:    &req.Object,
		SubjectId: &req.TenantRelation.Identifier,
	}
}

func (c *Client) toKetoCreateRelationshipBody(req ucasetypes.CreateRelationTupleRequest) keto.CreateRelationshipBody {
	return keto.CreateRelationshipBody{
		Namespace: &req.Namespace,
		Relation:  &req.Relation,
		Object:    &req.Object,

		SubjectId: &req.GlobalUserID,
	}
}
