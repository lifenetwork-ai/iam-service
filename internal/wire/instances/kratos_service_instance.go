package instances

import (
	"sync"

	"github.com/lifenetwork-ai/iam-service/internal/adapters/services/kratos"
	domainrepo "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/repositories"
	domainservice "github.com/lifenetwork-ai/iam-service/internal/domain/ucases/services"
)

var (
	kratosOnce     sync.Once
	kratosInstance domainservice.KratosService
)

func KratosServiceInstance(tenantRepo domainrepo.TenantRepository) domainservice.KratosService {
	kratosOnce.Do(func() {
		kratosInstance = kratos.NewKratosService(tenantRepo)
	})
	return kratosInstance
}
