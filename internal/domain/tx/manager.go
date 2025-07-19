package tx

import "context"

// Manager defines a contract for handling transactional operations within the given context.
type Manager interface {
	Transact(ctx context.Context, fn func(context.Context) error) error
}
