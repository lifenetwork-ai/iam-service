package ucases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/lifenetwork-ai/iam-service/constants"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	mock_repositories "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/repositories"
	mock_services "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_rate_limiter "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
)

func TestRegister_OrphanDeleteFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	identityID := "identity-123"
	tenantID := uuid.New()

	rateLimiter := mock_rate_limiter.NewMockRateLimiter(ctrl)
	rateLimiter.EXPECT().IsLimited(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
	rateLimiter.EXPECT().RegisterAttempt(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tenantRepo := mock_repositories.NewMockTenantRepository(ctrl)
	tenantRepo.EXPECT().GetByID(tenantID).Times(0)

	identityRepo := mock_repositories.NewMockUserIdentityRepository(ctrl)
	identityRepo.EXPECT().ExistsWithinTenant(ctx, tenantID.String(), constants.IdentifierEmail.String(), "test@example.com").Return(true, nil)
	identityRepo.EXPECT().GetByTypeAndValue(ctx, nil, tenantID.String(), constants.IdentifierEmail.String(), "test@example.com").Return(&domain.UserIdentity{ID: identityID, TenantID: tenantID.String(), KratosUserID: uuid.NewString()}, nil)
	identityRepo.EXPECT().Delete(nil, identityID).Return(assert.AnError)

	kratos := mock_services.NewMockKratosService(ctrl)
	kratos.EXPECT().GetIdentity(ctx, tenantID, gomock.Any()).Return(nil, errors.New("identity missing"))

	u := &userUseCase{
		rateLimiter:      rateLimiter,
		tenantRepo:       tenantRepo,
		userIdentityRepo: identityRepo,
		kratosService:    kratos,
	}

	resp, derr := u.Register(ctx, tenantID, "en", "test@example.com", "")
	assert.Nil(t, resp)
	assert.NotNil(t, derr)
	assert.Equal(t, "MSG_DELETE_IDENTIFIER_FAILED", derr.Code)
}
