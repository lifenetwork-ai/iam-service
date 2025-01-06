package interfaces

import "github.com/genefriendway/human-network-auth/internal/domain"

type DataAccessRequestRepository interface {
	CreateDataAccessRequest(request *domain.DataAccessRequest) error
}
