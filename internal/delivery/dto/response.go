package dto

type PaginationDTOResponse struct {
	NextPage int           `json:"next_page"`
	Page     int           `json:"page"`
	Size     int           `json:"size"`
	Total    int64         `json:"total,omitempty"`
	Data     []interface{} `json:"data"`
}

type ErrorDTOResponse struct {
	Status  int           `json:"status,omitempty"`
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details,omitempty"`
}
