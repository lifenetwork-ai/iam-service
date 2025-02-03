package interfaces

import "github.com/genefriendway/human-network-iam/internal/domain"

type PermissionRepository interface {
	PermissionExists(policyID, resource, action string) (bool, error)
	CreatePermission(permission *domain.Permission) error
}
