package search

type Result[T any] struct {
	TotalRows uint64 `json:"total_rows"`
	Rows      []*T   `json:"rows"`
}

type Request struct {
	Orders  []Order `json:"orders,omitempty"`
	Filters Filters `json:"filters,omitempty"`
	Limit   uint64  `json:"limit,omitempty" validate:"min=0"`
	Offset  uint64  `json:"offset,omitempty" validate:"min=0"`
}

type Filters map[string]any

type Order struct {
	Key  string `json:"key"`
	Desc bool   `json:"desc"`
}
