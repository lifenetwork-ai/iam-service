package ucases

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	dto "github.com/lifenetwork-ai/iam-service/internal/delivery/dto"
	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	mock_repositories "github.com/lifenetwork-ai/iam-service/mocks/adapters/repositories/types"
	"github.com/stretchr/testify/assert"
)

const (
	ValidOrgID      = "valid-org-id"
	InvalidOrgID    = "invalid-org-id"
	ValidOrgCode    = "VALID_ORG"
	InvalidOrgCode  = "INVALID_ORG"
	ValidOrgName    = "Valid Organization"
	ValidParentID   = "valid-parent-id"
	InvalidParentID = "invalid-parent-id"
)

var (
	ValidOrganization = entities.IdentityOrganization{
		ID:          ValidOrgID,
		Name:        ValidOrgName,
		Code:        ValidOrgCode,
		Description: "Valid organization description",
	}

	ValidParentOrganization = entities.IdentityOrganization{
		ID:          ValidParentID,
		Name:        "Parent Organization",
		Code:        "PARENT_ORG",
		Description: "Parent organization description",
	}
)

func TestOrganizationUseCase_List(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		size          int
		keyword       string
		mockOrgs      []entities.IdentityOrganization
		mockError     error
		expectedError *dto.ErrorDTOResponse
	}{
		{
			name:     "Success - List organizations",
			page:     1,
			size:     10,
			keyword:  "",
			mockOrgs: []entities.IdentityOrganization{ValidOrganization},
		},
		{
			name:     "Success - List organizations with keyword",
			page:     1,
			size:     10,
			keyword:  "valid",
			mockOrgs: []entities.IdentityOrganization{ValidOrganization},
		},
		{
			name:      "Error - Repository error",
			page:      1,
			size:      10,
			keyword:   "",
			mockError: errors.New("database error"),
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "database error",
				Details: []interface{}{"database error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrgRepo := mock_repositories.NewMockIdentityOrganizationRepository(ctrl)
			useCase := NewIdentityOrganizationUseCase(mockOrgRepo)

			ctx := context.Background()

			mockOrgRepo.EXPECT().
				Get(ctx, tt.size+1, (tt.page-1)*tt.size, tt.keyword).
				Return(tt.mockOrgs, tt.mockError)

			result, err := useCase.List(ctx, tt.page, tt.size, tt.keyword)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Status, err.Status)
				assert.Equal(t, tt.expectedError.Message, err.Message)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.page, result.Page)
				assert.Equal(t, tt.size, result.Size)
				assert.Len(t, result.Data, len(tt.mockOrgs))
			}
		})
	}
}

func TestOrganizationUseCase_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		orgID         string
		mockOrg       *entities.IdentityOrganization
		mockError     error
		expectedError *dto.ErrorDTOResponse
	}{
		{
			name:    "Success - Get organization by ID",
			orgID:   ValidOrgID,
			mockOrg: &ValidOrganization,
		},
		{
			name:    "Error - Organization not found",
			orgID:   InvalidOrgID,
			mockOrg: nil,
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusNotFound,
				Code:    "MSG_ORGANIZATION_NOT_FOUND",
				Message: "Organization not found",
				Details: []interface{}{
					map[string]string{"field": "id", "error": "Organization not found"},
				},
			},
		},
		{
			name:      "Error - Repository error",
			orgID:     ValidOrgID,
			mockError: errors.New("database error"),
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "database error",
				Details: []interface{}{"database error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrgRepo := mock_repositories.NewMockIdentityOrganizationRepository(ctrl)
			useCase := NewIdentityOrganizationUseCase(mockOrgRepo)

			ctx := context.Background()

			mockOrgRepo.EXPECT().
				GetByID(ctx, tt.orgID).
				Return(tt.mockOrg, tt.mockError)

			result, err := useCase.GetByID(ctx, tt.orgID)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Status, err.Status)
				assert.Equal(t, tt.expectedError.Message, err.Message)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockOrg.ID, result.ID)
				assert.Equal(t, tt.mockOrg.Name, result.Name)
				assert.Equal(t, tt.mockOrg.Code, result.Code)
			}
		})
	}
}

func TestOrganizationUseCase_Create(t *testing.T) {
	tests := []struct {
		name          string
		payload       dto.CreateIdentityOrganizationPayloadDTO
		mockExisting  *entities.IdentityOrganization
		mockParent    *entities.IdentityOrganization
		mockError     error
		expectedError *dto.ErrorDTOResponse
	}{
		{
			name: "Success - Create organization without parent",
			payload: dto.CreateIdentityOrganizationPayloadDTO{
				Name:        ValidOrgName,
				Code:        ValidOrgCode,
				Description: "Test organization",
			},
		},
		{
			name: "Success - Create organization with parent",
			payload: dto.CreateIdentityOrganizationPayloadDTO{
				Name:        ValidOrgName,
				Code:        ValidOrgCode,
				Description: "Test organization",
				ParentID:    ValidParentID,
			},
			mockParent: &ValidParentOrganization,
		},
		{
			name: "Error - Organization code already exists",
			payload: dto.CreateIdentityOrganizationPayloadDTO{
				Name:        ValidOrgName,
				Code:        ValidOrgCode,
				Description: "Test organization",
			},
			mockExisting: &ValidOrganization,
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "Organization with code VALID_ORG already exists",
				Details: []interface{}{"Organization with code VALID_ORG already exists"},
			},
		},
		{
			name: "Error - Parent organization not found",
			payload: dto.CreateIdentityOrganizationPayloadDTO{
				Name:        ValidOrgName,
				Code:        ValidOrgCode,
				Description: "Test organization",
				ParentID:    InvalidParentID,
			},
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "Parent organization with ID invalid-parent-id does not exist",
				Details: []interface{}{"Parent organization with ID invalid-parent-id does not exist"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrgRepo := mock_repositories.NewMockIdentityOrganizationRepository(ctrl)
			useCase := NewIdentityOrganizationUseCase(mockOrgRepo)

			ctx := context.Background()

			mockOrgRepo.EXPECT().
				GetByCode(ctx, tt.payload.Code).
				Return(tt.mockExisting, nil)

			if tt.mockExisting == nil && tt.payload.ParentID != "" {
				mockOrgRepo.EXPECT().
					GetByID(ctx, tt.payload.ParentID).
					Return(tt.mockParent, nil)
			}

			if tt.expectedError == nil {
				mockOrgRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(&ValidOrganization, nil)
			}

			result, err := useCase.Create(ctx, tt.payload)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Status, err.Status)
				assert.Equal(t, tt.expectedError.Message, err.Message)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, ValidOrganization.ID, result.ID)
				assert.Equal(t, ValidOrganization.Name, result.Name)
				assert.Equal(t, ValidOrganization.Code, result.Code)
			}
		})
	}
}

func TestOrganizationUseCase_Update(t *testing.T) {
	tests := []struct {
		name          string
		orgID         string
		payload       dto.UpdateIdentityOrganizationPayloadDTO
		mockOrg       *entities.IdentityOrganization
		mockExisting  *entities.IdentityOrganization
		mockParent    *entities.IdentityOrganization
		mockError     error
		expectedError *dto.ErrorDTOResponse
	}{
		{
			name:  "Success - Update organization name",
			orgID: ValidOrgID,
			payload: dto.UpdateIdentityOrganizationPayloadDTO{
				Name: "Updated Organization",
			},
			mockOrg: &ValidOrganization,
		},
		{
			name:  "Success - Update organization code",
			orgID: ValidOrgID,
			payload: dto.UpdateIdentityOrganizationPayloadDTO{
				Code: "UPDATED_ORG",
			},
			mockOrg: &ValidOrganization,
		},
		{
			name:  "Success - Update organization with parent",
			orgID: ValidOrgID,
			payload: dto.UpdateIdentityOrganizationPayloadDTO{
				ParentID: ValidParentID,
			},
			mockOrg:    &ValidOrganization,
			mockParent: &ValidParentOrganization,
		},
		{
			name:  "Error - Organization not found",
			orgID: InvalidOrgID,
			payload: dto.UpdateIdentityOrganizationPayloadDTO{
				Name: "Updated Organization",
			},
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "Organization with ID invalid-org-id does not exist",
				Details: []interface{}{"Organization with ID invalid-org-id does not exist"},
			},
		},
		{
			name:  "Error - Code already exists",
			orgID: ValidOrgID,
			payload: dto.UpdateIdentityOrganizationPayloadDTO{
				Code: "EXISTING_ORG",
			},
			mockOrg:      &ValidOrganization,
			mockExisting: &ValidOrganization,
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "Organization with code EXISTING_ORG already exists",
				Details: []interface{}{"Organization with code EXISTING_ORG already exists"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrgRepo := mock_repositories.NewMockIdentityOrganizationRepository(ctrl)
			useCase := NewIdentityOrganizationUseCase(mockOrgRepo)

			ctx := context.Background()

			mockOrgRepo.EXPECT().
				GetByID(ctx, tt.orgID).
				Return(tt.mockOrg, nil)

			if tt.mockOrg != nil && tt.payload.Code != "" {
				mockOrgRepo.EXPECT().
					GetByCode(ctx, tt.payload.Code).
					Return(tt.mockExisting, nil)
			}

			if tt.mockOrg != nil && tt.payload.ParentID != "" {
				mockOrgRepo.EXPECT().
					GetByID(ctx, tt.payload.ParentID).
					Return(tt.mockParent, nil)
			}

			if tt.expectedError == nil {
				mockOrgRepo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(&ValidOrganization, nil)
			}

			result, err := useCase.Update(ctx, tt.orgID, tt.payload)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Status, err.Status)
				assert.Equal(t, tt.expectedError.Message, err.Message)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, ValidOrganization.ID, result.ID)
				assert.Equal(t, ValidOrganization.Name, result.Name)
				assert.Equal(t, ValidOrganization.Code, result.Code)
			}
		})
	}
}

func TestOrganizationUseCase_Delete(t *testing.T) {
	tests := []struct {
		name          string
		orgID         string
		mockOrg       *entities.IdentityOrganization
		mockError     error
		expectedError *dto.ErrorDTOResponse
	}{
		{
			name:    "Success - Delete organization",
			orgID:   ValidOrgID,
			mockOrg: &ValidOrganization,
		},
		{
			name:  "Error - Organization not found",
			orgID: InvalidOrgID,
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "Organization with ID invalid-org-id does not exist",
				Details: []interface{}{"Organization with ID invalid-org-id does not exist"},
			},
		},
		{
			name:      "Error - Repository error",
			orgID:     ValidOrgID,
			mockError: errors.New("database error"),
			expectedError: &dto.ErrorDTOResponse{
				Status:  http.StatusBadRequest,
				Message: "database error",
				Details: []interface{}{"database error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrgRepo := mock_repositories.NewMockIdentityOrganizationRepository(ctrl)
			useCase := NewIdentityOrganizationUseCase(mockOrgRepo)

			ctx := context.Background()

			mockOrgRepo.EXPECT().
				GetByID(ctx, tt.orgID).
				Return(tt.mockOrg, tt.mockError)

			if tt.mockOrg != nil && tt.mockError == nil {
				mockOrgRepo.EXPECT().
					SoftDelete(ctx, tt.orgID).
					Return(&ValidOrganization, nil)
			}

			result, err := useCase.Delete(ctx, tt.orgID)

			if tt.expectedError != nil {
				assert.NotNil(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError.Status, err.Status)
				assert.Equal(t, tt.expectedError.Message, err.Message)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, ValidOrganization.ID, result.ID)
				assert.Equal(t, ValidOrganization.Name, result.Name)
				assert.Equal(t, ValidOrganization.Code, result.Code)
			}
		})
	}
}
