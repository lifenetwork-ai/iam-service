package keto

import (
	"context"
	"net/http"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	repotypes "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	ucasetypes "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	keto "github.com/ory/keto-client-go"
)

// Client wraps the Keto client configuration
type Client struct {
	tenantRepo repotypes.TenantRepository
	client     *keto.APIClient
	config     *conf.KetoConfiguration
}

var _ ucases.KetoService = (*Client)(nil)

// NewClient creates a new Keto client
func NewKetoService(tenantRepo repotypes.TenantRepository) ucases.KetoService {
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
	req := c.client.PermissionApi.PostCheckPermission(ctx).PostCheckPermissionBody(request.ToKetoPostCheckPermissionBody())
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

// // BatchCheckPermission checks if a subject has permission to perform an action on an object
// func (c *Client) BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, error) {
// 	// Create the request body
// 	type requestBody struct {
// 		Namespace  string           `json:"namespace"`
// 		Object     string           `json:"object"`
// 		Relation   string           `json:"relation"`
// 	}

// 	var requests []requestBody
// 	for _, tuple := range dto.Tuples {
// 		req := requestBody{
// 			Namespace: tuple.Namespace,
// 			Object:    tuple.Object,
// 			Relation:  tuple.Relation,
// 		}
// 		requests = append(requests, req)
// 	}

// 	// Marshal the request body
// 	reqBody, err := json.Marshal(requests)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to marshal request body: %w", err)
// 	}

// 	// Parse the URL
// 	url, err := url.Parse(fmt.Sprintf("%s%s", c.config.DefaultReadURL, constants.BatchPermissionCheckEndpoint))
// 	if err != nil {
// 		return false, fmt.Errorf("failed to parse URL: %w", err)
// 	}

// 	// Create the HTTP request
// 	httpReq, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewBuffer(reqBody))
// 	if err != nil {
// 		return false, fmt.Errorf("failed to create request: %w", err)
// 	}
// 	httpReq.Header.Set("Content-Type", constants.ContentTypeJson)

// 	// Send the request
// 	client := &http.Client{}
// 	resp, err := client.Do(httpReq)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to send batch permission check request: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return false, fmt.Errorf("batch permission check failed with status: %s", resp.Status)
// 	}

// 	// Parse response
// 	type responseBody struct {
// 		Allowed bool `json:"allowed"`
// 	}

// 	var responses []responseBody
// 	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
// 		return false, fmt.Errorf("failed to decode batch permission check response: %w", err)
// 	}

// 	// Check if all permissions are allowed
// 	for _, response := range responses {
// 		if !response.Allowed {
// 			return false, nil
// 		}
// 	}

// 	return true, nil
// }

// CreateRelationTuple creates a relation tuple
// Note: The dto should be validated before calling this function
func (c *Client) CreateRelationTuple(ctx context.Context, request ucasetypes.CreateRelationTupleRequest) *domainerrors.DomainError {
	logger.GetLogger().Debugf("Creating relation tuple for namespace: %s, object: %s, relation: %s, subject_set: %v",
		request.Namespace, request.Object, request.Relation, request.SubjectSet)

	req := c.client.RelationshipApi.CreateRelationship(ctx).CreateRelationshipBody(request.ToKetoCreateRelationshipBody())

	// Log the request details before execution
	logger.GetLogger().Debugf("Sending request to Keto Write API URL: %s", c.config.DefaultWriteURL)
	logger.GetLogger().Debugf("Request body: %+v", request.ToKetoCreateRelationshipBody())

	_, httpResp, err := req.Execute()
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
