package dto

import "fmt"

type PaginationDTOResponse[T any] struct {
	NextPage   int   `json:"next_page"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	Items      []T   `json:"items"`
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
