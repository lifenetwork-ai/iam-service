package dto

import "fmt"

// PaginationDTOResponse is a generic response for pagination
// It is used to return a paginated list of items
type PaginationDTOResponse[T any] struct {
	NextPage   int   `json:"next_page"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	Items      []T   `json:"items"`
}

// TenantPaginationDTOResponse is a concrete response for tenant pagination
// This is used specifically for swagger documentation compatibility
type TenantPaginationDTOResponse struct {
	NextPage   int         `json:"next_page"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int64       `json:"total_count"`
	Items      []TenantDTO `json:"items"`
}

type SuccessDTOResponse struct {
	Status  int         `json:"status,omitempty"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorDTOResponse struct {
	Status  int           `json:"status,omitempty"`
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details,omitempty"`
}

func (e *ErrorDTOResponse) Error() string {
	return fmt.Sprintf("code: %s, message: %s, details: %v", e.Code, e.Message, e.Details)
}

// CheckIdentifierResponse represents the response for checking if an identifier exists.
type CheckIdentifierResponse struct {
	Registered bool `json:"registered"`
}

type AdminAddIdentifierResponse struct {
	GlobalUserID string `json:"global_user_id"`
	KratosUserID string `json:"kratos_user_id"`
	Identifier   string `json:"new_identifier"`
	Lang         string `json:"lang"`
}
