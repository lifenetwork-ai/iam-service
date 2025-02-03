package dto

// OrganizationDTO represents an organization.
type OrganizationDTO struct {
	ID 		string `json:"id"`
	Name 	string `json:"name"`
	Code 	string `json:"code"`
	OwnerID string `json:"owner_id"`
}
