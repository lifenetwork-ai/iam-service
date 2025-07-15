package keto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lifenetwork-ai/iam-service/conf"
	repo_types "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories/types"
	"github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	keto "github.com/ory/keto-client-go"
)

// Client wraps the Keto client configuration
type Client struct {
	tenantRepo repo_types.TenantRepository
	client     *keto.APIClient
	config     *conf.KetoConfiguration
}

// NewClient creates a new Keto client
func NewClient(cfg *conf.KetoConfiguration, tenantRepo repo_types.TenantRepository) *Client {
	ketoCfg := keto.NewConfiguration()
	// Set the server URL from the configuration
	ketoCfg.Servers = keto.ServerConfigurations{
		{
			URL:         cfg.DefaultReadURL,
			Description: "Keto Read API",
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
func (c *Client) CheckPermission(ctx context.Context, dto dto.CheckPermissionRequestDTO) (bool, error) {
	req := c.client.PermissionApi.PostCheckPermission(ctx).PostCheckPermissionBody(keto.PostCheckPermissionBody{
		Namespace: &dto.Namespace,
		Relation:  &dto.Action,
		Object:    &dto.Object,
		SubjectId: &dto.SubjectID,
		SubjectSet: &keto.SubjectSet{
			Namespace: dto.SubjectSet.Namespace,
			Relation:  dto.SubjectSet.Relation,
			Object:    dto.SubjectSet.Object,
		},
	})
	ketoResp, _, err := req.Execute()
	if err != nil {
		return false, err
	}

	return ketoResp.GetAllowed(), nil
}

// BatchCheckPermission checks if a subject has permission to perform an action on an object
func (c *Client) BatchCheckPermission(ctx context.Context, dto dto.BatchCheckPermissionRequestDTO) (bool, error) {
	// Create the request body
	type requestBody struct {
		Namespace  string           `json:"namespace"`
		Object     string           `json:"object"`
		Relation   string           `json:"relation"`
		SubjectID  string           `json:"subject_id,omitempty"`
		SubjectSet *keto.SubjectSet `json:"subject_set,omitempty"`
	}

	var requests []requestBody
	for _, tuple := range dto.Tuples {
		req := requestBody{
			Namespace: tuple.Namespace,
			Object:    tuple.Object,
			Relation:  tuple.Action,
		}

		if tuple.SubjectID != "" {
			req.SubjectID = tuple.SubjectID
		}

		if tuple.SubjectSet != nil {
			req.SubjectSet = &keto.SubjectSet{
				Namespace: tuple.SubjectSet.Namespace,
				Relation:  tuple.SubjectSet.Relation,
				Object:    tuple.SubjectSet.Object,
			}
		}

		requests = append(requests, req)
	}

	// Marshal the request body
	reqBody, err := json.Marshal(requests)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Parse the URL
	url, err := url.Parse(fmt.Sprintf("%s/check/permission/bulk", c.config.DefaultReadURL))
	if err != nil {
		return false, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to send batch permission check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("batch permission check failed with status: %s", resp.Status)
	}

	// Parse response
	type responseBody struct {
		Allowed bool `json:"allowed"`
	}

	var responses []responseBody
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return false, fmt.Errorf("failed to decode batch permission check response: %w", err)
	}

	// Check if all permissions are allowed
	for _, response := range responses {
		if !response.Allowed {
			return false, nil
		}
	}

	return true, nil
}

// CreateRelationTuple creates a relation tuple
func (c *Client) CreateRelationTuple(ctx context.Context, dto dto.CreateRelationTupleRequestDTO) error {
	req := c.client.RelationshipApi.CreateRelationship(ctx).CreateRelationshipBody(keto.CreateRelationshipBody{
		Namespace: &dto.Namespace,
		Relation:  &dto.Relation,
		Object:    &dto.Object,
		SubjectId: &dto.SubjectID,
		SubjectSet: &keto.SubjectSet{
			Namespace: dto.SubjectSet.Namespace,
		},
	})
	_, httpResp, err := req.Execute()
	if err != nil {
		return err
	}

	if httpResp.StatusCode != http.StatusOK {
		logger.GetLogger().Errorf("failed to create relation tuple: %s", httpResp.Status)
		return fmt.Errorf("failed to create relation tuple: %s", httpResp.Status)
	}

	return nil
}
