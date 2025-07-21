package types

type PaginatedResponse[T any] struct {
	Items      []T
	TotalCount int64
	Page       int
	PageSize   int
	NextPage   int
}

func CalculatePagination[T any](items []T, page, size int) *PaginatedResponse[T] {
	start := (page - 1) * size
	end := start + size

	if start >= len(items) {
		start = len(items)
	}

	if end > len(items) {
		end = len(items)
	}

	nextPage := page
	if end < len(items) {
		nextPage++
	}

	return &PaginatedResponse[T]{
		Items:      items[start:end],
		TotalCount: int64(len(items)),
		Page:       page,
		PageSize:   size,
		NextPage:   nextPage,
	}
}
