//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq" // postgres driver for wait.ForSQL
	"go.uber.org/mock/gomock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/lifenetwork-ai/iam-service/infrastructures/caching"
	adaptersrepo "github.com/lifenetwork-ai/iam-service/internal/adapters/repositories"
	kratos_service "github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domain "github.com/lifenetwork-ai/iam-service/internal/domain/entities"
	ucases "github.com/lifenetwork-ai/iam-service/internal/domain/ucases"
	"github.com/lifenetwork-ai/iam-service/internal/domain/ucases/interfaces"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	domainservice "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
	mock_cache_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/caching/types"
	mock_rl_types "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/rate_limiter/types"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testDeps groups repositories and fakes used by tests
type testDeps struct {
	cacheRepo                 *mock_cache_types.MockCacheRepository
	tenantRepo                domainrepo.TenantRepository
	globalUserRepo            domainrepo.GlobalUserRepository
	userIdentityRepo          domainrepo.UserIdentityRepository
	userIdentifierMappingRepo domainrepo.UserIdentifierMappingRepository
	challengeSessionRepo      domainrepo.ChallengeSessionRepository
	kratosService             domainservice.KratosService
	rateLimiter               *mock_rl_types.MockRateLimiter
}

// containerSem limits how many Postgres containers run concurrently.
var containerSem = make(chan struct{}, getMaxContainers())

func getMaxContainers() int {
	// Default to 2, override with TEST_MAX_CONTAINERS (>0)
	if v := os.Getenv("TEST_MAX_CONTAINERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 2
}

func applyPostgresScriptsFromTestFile(t require.TestingT, db *gorm.DB) {
	_, thisFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(thisFile)
	scriptsDir := filepath.Join(testDir, "../../../adapters/postgres/scripts")
	entries, err := os.ReadDir(scriptsDir)
	require.NoError(t, err, "read scripts dir")
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	for _, name := range files {
		path := filepath.Join(scriptsDir, name)
		content, err := os.ReadFile(path)
		require.NoError(t, err, "read %s", name)
		require.NoError(t, db.Exec(string(content)).Error, "execute %s", name)
	}
}

func startPostgresAndBuildUCase(t *testing.T, ctx context.Context, ctrl *gomock.Controller, seedTenantName string) (interfaces.IdentityUserUseCase, interfaces.AdminUseCase, testDeps, *gorm.DB, uuid.UUID, testcontainers.Container) {
	// Acquire slot for container; release when test finishes
	containerSem <- struct{}{}
	t.Cleanup(func() { <-containerSem })
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(45 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("host=%s user=postgres password=postgres dbname=testdb port=%s sslmode=disable TimeZone=UTC", host, port.Port())
	var db *gorm.DB
	// Retry DB connection to avoid race where port is listening but DB not ready
	for i := 0; i < 20; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			if sqlDB, derr := db.DB(); derr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					break
				}
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	require.NoError(t, err)

	_ = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error
	_ = db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";").Error
	applyPostgresScriptsFromTestFile(t, db)

	inMemCache := caching.NewCachingRepository(context.Background(), caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)))
	deps := testDeps{}
	deps.cacheRepo = mock_cache_types.NewMockCacheRepository(ctrl)
	deps.tenantRepo = adaptersrepo.NewTenantRepository(db)
	deps.globalUserRepo = adaptersrepo.NewGlobalUserRepository(db)
	deps.userIdentityRepo = adaptersrepo.NewUserIdentityRepository(db)
	deps.userIdentifierMappingRepo = adaptersrepo.NewUserIdentifierMappingRepository(db)
	deps.challengeSessionRepo = adaptersrepo.NewChallengeSessionRepository(inMemCache)
	deps.kratosService = kratos_service.NewFakeKratosService()
	deps.rateLimiter = mock_rl_types.NewMockRateLimiter(ctrl)

	ucase := ucases.NewIdentityUserUseCase(
		db,
		deps.cacheRepo,
		deps.rateLimiter,
		deps.challengeSessionRepo,
		deps.tenantRepo,
		deps.globalUserRepo,
		deps.userIdentityRepo,
		deps.userIdentifierMappingRepo,
		deps.kratosService,
	)

	// Create admin use case
	adminUcase := ucases.NewAdminUseCase(
		deps.tenantRepo,
		adaptersrepo.NewAdminAccountRepository(db),
		deps.userIdentityRepo,
		deps.userIdentifierMappingRepo,
		deps.kratosService,
	)
	tenantID := uuid.New()
	require.NoError(t, db.Create(&domain.Tenant{ID: tenantID, Name: seedTenantName}).Error)

	return ucase, adminUcase, deps, db, tenantID, container
}

func startPostgresAndBuildAdminUCase(t *testing.T, ctx context.Context, ctrl *gomock.Controller, seedTenantName string) (interfaces.AdminUseCase, testDeps, *gorm.DB, uuid.UUID, testcontainers.Container) {
	// Acquire slot for container; release when test finishes
	containerSem <- struct{}{}
	t.Cleanup(func() { <-containerSem })
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(45 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := fmt.Sprintf("host=%s user=postgres password=postgres dbname=testdb port=%s sslmode=disable TimeZone=UTC", host, port.Port())
	var db *gorm.DB
	// Retry DB connection to avoid race where port is listening but DB not ready
	for i := 0; i < 20; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			if sqlDB, derr := db.DB(); derr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					break
				}
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	require.NoError(t, err)

	_ = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error
	_ = db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";").Error
	applyPostgresScriptsFromTestFile(t, db)

	inMemCache := caching.NewCachingRepository(context.Background(), caching.NewGoCacheClient(cache.New(5*time.Minute, 10*time.Minute)))
	deps := testDeps{}
	deps.tenantRepo = adaptersrepo.NewTenantRepository(db)
	deps.globalUserRepo = adaptersrepo.NewGlobalUserRepository(db)
	deps.userIdentityRepo = adaptersrepo.NewUserIdentityRepository(db)
	deps.userIdentifierMappingRepo = adaptersrepo.NewUserIdentifierMappingRepository(db)
	deps.challengeSessionRepo = adaptersrepo.NewChallengeSessionRepository(inMemCache)
	deps.kratosService = kratos_service.NewFakeKratosService()
	deps.rateLimiter = mock_rl_types.NewMockRateLimiter(ctrl)

	// Create admin use case
	adminUcase := ucases.NewAdminUseCase(
		deps.tenantRepo,
		adaptersrepo.NewAdminAccountRepository(db),
		deps.userIdentityRepo,
		deps.userIdentifierMappingRepo,
		deps.kratosService,
	)
	tenantID := uuid.New()
	require.NoError(t, db.Create(&domain.Tenant{ID: tenantID, Name: seedTenantName}).Error)

	return adminUcase, deps, db, tenantID, container
}
