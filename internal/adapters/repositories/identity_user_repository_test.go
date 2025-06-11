package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	entities "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
)

// MockCacheRepository implements infrainterfaces.CacheRepository for testing
type MockCacheRepository struct {
	items map[string]interface{}
}

func NewMockCacheRepository() *MockCacheRepository {
	return &MockCacheRepository{
		items: make(map[string]interface{}),
	}
}

func (m *MockCacheRepository) SaveItem(key fmt.Stringer, value interface{}, ttl time.Duration) error {
	m.items[key.String()] = value
	return nil
}

func (m *MockCacheRepository) RetrieveItem(key fmt.Stringer, value interface{}) error {
	if item, exists := m.items[key.String()]; exists {
		value = item
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (m *MockCacheRepository) RemoveItem(key fmt.Stringer) error {
	delete(m.items, key.String())
	return nil
}

// TestIdentityUser is a simplified version of IdentityUser for testing
type TestIdentityUser struct {
	ID             string `gorm:"primaryKey"`
	Seed           string
	OrganizationId string    `gorm:"column:organization_id;not null"`
	UserName       string    `gorm:"column:user_name;not null"`
	Email          string    `gorm:"column:email;not null"`
	Phone          string    `gorm:"column:phone;not null"`
	PasswordHash   string    `gorm:"column:password_hash;not null"`
	Status         bool      `gorm:"column:status;not null;default:true"`
	Name           string    `gorm:"column:name;not null"`
	FirstName      string    `gorm:"column:first_name"`
	LastName       string    `gorm:"column:last_name"`
	FullName       string    `gorm:"column:full_name"`
	LastLoginAt    time.Time `gorm:"column:last_login_at"`

	SelfAuthenticateID string `gorm:"column:self_authenticate_id"`
	GoogleID           string `gorm:"column:google_id"`
	FacebookID         string `gorm:"column:facebook_id"`
	AppleID            string `gorm:"column:apple_id"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (m *TestIdentityUser) TableName() string {
	return "identity_users"
}

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true,
	})
	require.NoError(t, err)

	// Migrate the schema using the test model
	err = db.AutoMigrate(&TestIdentityUser{})
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.Close()
	}

	return db, cleanup
}

func TestIdentityUserRepository_CRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mockCache := NewMockCacheRepository()
	repo := NewIdentityUserRepository(db, mockCache)

	ctx := context.WithValue(context.Background(), "organizationId", "test-org-id")
	ctx = context.WithValue(ctx, "organization", &entities.IdentityOrganization{
		Code: "TEST",
	})

	t.Run("Create and FindByID", func(t *testing.T) {
		// Create test user
		user := &entities.IdentityUser{
			ID:           uuid.New().String(),
			UserName:     "testuser",
			Email:        "test@example.com",
			Phone:        "1234567890",
			PasswordHash: "hashedpassword",
			Name:         "Test User",
			FirstName:    "Test",
			LastName:     "User",
			FullName:     "Test User",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.NotEmpty(t, user.Seed)

		// Find by ID
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.UserName, found.UserName)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("FindByEmail", func(t *testing.T) {
		user := &entities.IdentityUser{
			ID:           uuid.New().String(),
			UserName:     "emailuser",
			Email:        "email@example.com",
			Phone:        "9876543210",
			PasswordHash: "hashedpassword",
			Name:         "Email User",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("FindByPhone", func(t *testing.T) {
		user := &entities.IdentityUser{
			ID:           uuid.New().String(),
			UserName:     "phoneuser",
			Email:        "phone@example.com",
			Phone:        "5555555555",
			PasswordHash: "hashedpassword",
			Name:         "Phone User",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByPhone(ctx, user.Phone)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.Phone, found.Phone)
	})

	t.Run("FindByUsername", func(t *testing.T) {
		user := &entities.IdentityUser{
			ID:           uuid.New().String(),
			UserName:     "usernameuser",
			Email:        "username@example.com",
			Phone:        "1111111111",
			PasswordHash: "hashedpassword",
			Name:         "Username User",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByUsername(ctx, user.UserName)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.UserName, found.UserName)
	})

	t.Run("Update", func(t *testing.T) {
		user := &entities.IdentityUser{
			ID:           uuid.New().String(),
			UserName:     "updateuser",
			Email:        "update@example.com",
			Phone:        "2222222222",
			PasswordHash: "hashedpassword",
			Name:         "Update User",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Update user
		user.Name = "Updated Name"
		user.Email = "updated@example.com"
		err = repo.Update(ctx, user)
		require.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "updated@example.com", found.Email)
	})

	t.Run("Delete", func(t *testing.T) {
		user := &entities.IdentityUser{
			ID:           uuid.New().String(),
			UserName:     "deleteuser",
			Email:        "delete@example.com",
			Phone:        "3333333333",
			PasswordHash: "hashedpassword",
			Name:         "Delete User",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Delete user
		err = repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		// Verify deletion
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("FindByExternalIDs", func(t *testing.T) {
		user := &entities.IdentityUser{
			ID:                 uuid.New().String(),
			UserName:           "externaluser",
			Email:              "external@example.com",
			Phone:              "4444444444",
			PasswordHash:       "hashedpassword",
			Name:               "External User",
			SelfAuthenticateID: "self-auth-123",
			GoogleID:           "google-123",
			FacebookID:         "facebook-123",
			AppleID:            "apple-123",
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Test FindBySelfAuthenticateID
		found, err := repo.FindBySelfAuthenticateID(ctx, user.SelfAuthenticateID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.SelfAuthenticateID, found.SelfAuthenticateID)

		// Test FindByGoogleID
		found, err = repo.FindByGoogleID(ctx, user.GoogleID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.GoogleID, found.GoogleID)

		// Test FindByFacebookID
		found, err = repo.FindByFacebookID(ctx, user.FacebookID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.FacebookID, found.FacebookID)

		// Test FindByAppleID
		found, err = repo.FindByAppleID(ctx, user.AppleID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.AppleID, found.AppleID)
	})
}
