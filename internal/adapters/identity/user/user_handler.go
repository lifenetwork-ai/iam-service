package identity_user

import (
	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type userHandler struct {
	ucase interfaces.IdentityUserUseCase
}

func NewIdentityUserHandler(ucase interfaces.IdentityUserUseCase) *userHandler {
	return &userHandler{
		ucase: ucase,
	}
}
