package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/human-network-iam/internal/interfaces"
)

type organizationHandler struct {
	organizationUCase  interfaces.OrganizationUseCase
}

func NewOrganizationHandler(organizationUCase interfaces.OrganizationUseCase) *organizationHandler {
	return &organizationHandler{
		organizationUCase:  organizationUCase,
	}
}

func (h *organizationHandler) GetOrganizations(c *gin.Context) {
	
}

func (h *organizationHandler) GetOrganizationByID(c *gin.Context) {
	
}

func (h *organizationHandler) CreateOrganization(c *gin.Context) {

}

func (h *organizationHandler) UpdateOrganization(c *gin.Context) {

}

func (h *organizationHandler) DeleteOrganization(c *gin.Context) {
	
}

func (h *organizationHandler) GetMembers(c *gin.Context) {
	
}

func (h *organizationHandler) AddMember(c *gin.Context) {
	
}

func (h *organizationHandler) RemoveMember(c *gin.Context) {
	
}
