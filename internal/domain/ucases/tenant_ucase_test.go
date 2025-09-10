package ucases

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	mock_repos "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	"github.com/stretchr/testify/assert"
)

func newTenantUC(ctrl *gomock.Controller) (*tenantUseCase, *mock_repos.MockTenantRepository) {
	repo := mock_repos.NewMockTenantRepository(ctrl)
	uc := NewTenantUseCase(repo).(*tenantUseCase)
	return uc, repo
}

func TestTenant_GetAll_Success_And_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, repo := newTenantUC(ctrl)

	repo.EXPECT().List().Return([]*domain.Tenant{{Name: "a"}}, nil)
	items, derr := uc.GetAll(context.Background())
	assert.Nil(t, derr)
	assert.Len(t, items, 1)

	repo.EXPECT().List().Return(nil, assert.AnError)
	items, derr = uc.GetAll(context.Background())
	assert.NotNil(t, derr)
	assert.Nil(t, items)
}

func TestTenant_List_Filter_Pagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, repo := newTenantUC(ctrl)

	t1 := &domain.Tenant{Name: "Alpha"}
	t2 := &domain.Tenant{Name: "Beta"}
	repo.EXPECT().List().Return([]*domain.Tenant{t1, t2}, nil).Times(2)

	// no filter
	resp, derr := uc.List(context.Background(), 1, 10, "")
	assert.Nil(t, derr)
	assert.Equal(t, int64(2), resp.TotalCount)

	// filter by alpha
	resp, derr = uc.List(context.Background(), 1, 10, "alp")
	assert.Nil(t, derr)
	assert.Equal(t, int64(1), resp.TotalCount)
}

func TestTenant_GetByID_Invalid_NotFound_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, repo := newTenantUC(ctrl)

	// invalid UUID
	item, derr := uc.GetByID(context.Background(), "not-uuid")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// not found (nil)
	repo.EXPECT().GetByID(gomock.Any()).Return(nil, nil)
	item, derr = uc.GetByID(context.Background(), "11111111-1111-1111-1111-111111111111")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// success
	repo.EXPECT().GetByID(gomock.Any()).Return(&domain.Tenant{Name: "Acme"}, nil)
	item, derr = uc.GetByID(context.Background(), "11111111-1111-1111-1111-111111111111")
	assert.Nil(t, derr)
	assert.Equal(t, "Acme", item.Name)
}

func TestTenant_Create_Conflict_And_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, repo := newTenantUC(ctrl)

	// conflict
	repo.EXPECT().GetByName("Acme").Return(&domain.Tenant{Name: "Acme"}, nil)
	item, derr := uc.Create(context.Background(), "Acme", "p", "a")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// success path
	repo2Ctrl := gomock.NewController(t)
	defer repo2Ctrl.Finish()
	uc2, repo2 := newTenantUC(repo2Ctrl)
	repo2.EXPECT().GetByName("New").Return(nil, nil)
	repo2.EXPECT().Create(gomock.AssignableToTypeOf(&domain.Tenant{})).Return(nil)
	item, derr = uc2.Create(context.Background(), "New", "p", "a")
	assert.Nil(t, derr)
	assert.Equal(t, "New", item.Name)
}

func TestTenant_Update_Paths(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, repo := newTenantUC(ctrl)

	// invalid uuid
	item, derr := uc.Update(context.Background(), "bad", "", "", "")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// not found
	repo.EXPECT().GetByID(gomock.Any()).Return(nil, nil)
	item, derr = uc.Update(context.Background(), "11111111-1111-1111-1111-111111111111", "", "", "")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// conflict on name update
	exist := &domain.Tenant{Name: "Old", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo.EXPECT().GetByID(gomock.Any()).Return(exist, nil)
	repo.EXPECT().GetByName("New").Return(&domain.Tenant{Name: "New"}, nil)
	item, derr = uc.Update(context.Background(), "11111111-1111-1111-1111-111111111111", "New", "", "")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// success update some fields -> Update called
	exist2 := &domain.Tenant{Name: "Old", PublicURL: "p1", AdminURL: "a1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo.EXPECT().GetByID(gomock.Any()).Return(exist2, nil)
	repo.EXPECT().GetByName("Old").Return(nil, nil).AnyTimes()
	repo.EXPECT().Update(gomock.AssignableToTypeOf(&domain.Tenant{})).Return(nil)
	item, derr = uc.Update(context.Background(), "11111111-1111-1111-1111-111111111111", "", "p2", "a2")
	assert.Nil(t, derr)
	assert.Equal(t, "p2", item.PublicURL)
	assert.Equal(t, "a2", item.AdminURL)
}

func TestTenant_Delete_Invalid_And_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, repo := newTenantUC(ctrl)

	// invalid
	item, derr := uc.Delete(context.Background(), "bad")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// not found
	repo.EXPECT().GetByID(gomock.Any()).Return(nil, nil)
	item, derr = uc.Delete(context.Background(), "11111111-1111-1111-1111-111111111111")
	assert.NotNil(t, derr)
	assert.Nil(t, item)

	// success
	tenant := &domain.Tenant{Name: "Acme"}
	repo.EXPECT().GetByID(gomock.Any()).Return(tenant, nil)
	repo.EXPECT().Delete(gomock.Any()).Return(nil)
	item, derr = uc.Delete(context.Background(), "11111111-1111-1111-1111-111111111111")
	assert.Nil(t, derr)
	assert.Equal(t, "Acme", item.Name)
}
