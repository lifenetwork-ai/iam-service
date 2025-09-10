package ucases

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	domainerrors "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/errors"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/types"
	mock_repos "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	"github.com/stretchr/testify/assert"
)

func newPermissionUC(ctrl *gomock.Controller) (*permissionUseCase, *mock_services.MockKetoService, *mock_repos.MockUserIdentityRepository) {
	keto := mock_services.NewMockKetoService(ctrl)
	repo := mock_repos.NewMockUserIdentityRepository(ctrl)
	uc := NewPermissionUseCase(keto, repo).(*permissionUseCase)
	return uc, keto, repo
}

func TestPermission_CheckPermission_Direct_And_Indirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, keto, repo := newPermissionUC(ctrl)
	ctx := context.Background()

	req := types.CheckPermissionRequest{
		Namespace:      "files",
		Object:         "files:doc1",
		Relation:       "view",
		TenantRelation: types.TenantRelation{TenantID: "tenant-1", Identifier: "user@example.com"},
	}

	// Resolve global user id
	repo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), "tenant-1", "email", "user@example.com").Return(&domain.UserIdentity{GlobalUserID: "g-1"}, nil).AnyTimes()

	// Direct allowed
	keto.EXPECT().CheckPermission(gomock.Any(), gomock.Any()).Return(true, nil)
	ok, derr := uc.CheckPermission(ctx, req)
	assert.Nil(t, derr)
	assert.True(t, ok)

	// Direct denied -> indirect via role
	keto.EXPECT().CheckPermission(gomock.Any(), gomock.Any()).Return(false, nil)
	keto.EXPECT().CheckPermission(gomock.Any(), gomock.Any()).Return(true, nil)
	ok, derr = uc.CheckPermission(ctx, req)
	assert.Nil(t, derr)
	assert.True(t, ok)
}

func TestPermission_CheckPermission_Validation_And_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, keto, repo := newPermissionUC(ctrl)
	ctx := context.Background()

	// invalid request -> Validate will fail (empty values)
	ok, derr := uc.CheckPermission(ctx, types.CheckPermissionRequest{})
	assert.NotNil(t, derr)
	assert.False(t, ok)

	// global user resolve error
	repo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), "tenant-1", "email", "user@example.com").Return(nil, assert.AnError)
	req := types.CheckPermissionRequest{Namespace: "files", Object: "files:doc1", Relation: "view", TenantRelation: types.TenantRelation{TenantID: "tenant-1", Identifier: "user@example.com"}}
	ok, derr = uc.CheckPermission(ctx, req)
	assert.NotNil(t, derr)
	assert.False(t, ok)

	// keto error on direct
	repo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), "tenant-1", "email", "user@example.com").Return(&domain.UserIdentity{GlobalUserID: "g-1"}, nil)
	keto.EXPECT().CheckPermission(gomock.Any(), gomock.Any()).Return(false, domainerrors.NewInternalError("X", "Y"))
	ok, derr = uc.CheckPermission(ctx, req)
	assert.NotNil(t, derr)
	assert.False(t, ok)
}

func TestPermission_DelegateAccess_Success_And_InvalidTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, keto, repo := newPermissionUC(ctrl)
	ctx := context.Background()

	repo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), "tenant-1", "email", "target@example.com").Return(&domain.UserIdentity{GlobalUserID: "tg-1"}, nil)
	// Expect role tuple then permission tuple
	keto.EXPECT().CreateRelationTuple(gomock.Any(), gomock.Any()).Return(nil)
	keto.EXPECT().CreateRelationTuple(gomock.Any(), gomock.Any()).Return(nil)
	req := types.DelegateAccessRequest{TenantID: "tenant-1", Identifier: "target@example.com", ResourceType: "files", ResourceID: "doc1", Permission: "manage"}
	ok, derr := uc.DelegateAccess(ctx, req)
	assert.Nil(t, derr)
	assert.True(t, ok)

	// invalid target (lookup error)
	repo2Ctrl := gomock.NewController(t)
	defer repo2Ctrl.Finish()
	uc2, keto2, repo2 := newPermissionUC(repo2Ctrl)
	repo2.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), "tenant-1", "email", "bad@example.com").Return(nil, assert.AnError)
	ok, derr = uc2.DelegateAccess(ctx, types.DelegateAccessRequest{TenantID: "tenant-1", Identifier: "bad@example.com", ResourceType: "files", ResourceID: "doc1", Permission: "view"})
	assert.NotNil(t, derr)
	assert.False(t, ok)
	_ = keto2 // silence unused in this subtest
}

func TestPermission_CreateRelationTuple_Success_And_Validation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, keto, repo := newPermissionUC(ctrl)
	ctx := context.Background()

	// validation fails
	derr := uc.CreateRelationTuple(ctx, types.CreateRelationTupleRequest{})
	assert.NotNil(t, derr)

	// success path
	repo.EXPECT().GetByTypeAndValue(gomock.Any(), gomock.Any(), "tenant-1", "email", "user@example.com").Return(&domain.UserIdentity{GlobalUserID: "g-1"}, nil)
	keto.EXPECT().CreateRelationTuple(gomock.Any(), gomock.Any()).Return(nil)
	derr = uc.CreateRelationTuple(ctx, types.CreateRelationTupleRequest{Namespace: "files", Relation: "edit", Object: "files:doc1", TenantRelation: types.TenantRelation{TenantID: "tenant-1", Identifier: "user@example.com"}})
	assert.Nil(t, derr)
}
