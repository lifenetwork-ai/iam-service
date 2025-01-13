package interfaces

import "github.com/genefriendway/human-network-auth/internal/domain"

type PermissionRepository interface {
	PermissionExists(policyID, resource, action string) (bool, error)
	CreatePermission(permission *domain.Permission) error
}
