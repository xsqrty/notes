package search

// Result represents a generic structure for paginated query results.
// TotalRows indicates the total number of rows available.
// Rows contain the actual data items returned, parameterized by the generic type T.
type Result[T any] struct {
	TotalRows uint64 `json:"total_rows"`
	Rows      []*T   `json:"rows"`
}

// Request represents the structure for search requests with orders, filters, and pagination parameters.
type Request struct {
	Orders  []Order `json:"orders,omitempty"`
	Filters Filters `json:"filters,omitempty"`
	Limit   uint64  `json:"limit,omitempty"   validate:"min=0"`
	Offset  uint64  `json:"offset,omitempty"  validate:"min=0"`
}

// Filters represent a set of nested key-value pairs used to filter data or queries dynamically.
type Filters map[string]any

// Order represents a sorting preference with a key and direction.
// Key specifies the field to sort by, and Desc defines whether sorting is descending.
type Order struct {
	Key  string `json:"key"`
	Desc bool   `json:"desc"`
}
