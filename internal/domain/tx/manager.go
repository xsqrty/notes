package tx

import "context"

type Manager interface {
	Transact(ctx context.Context, fn func(context.Context) error) error
}
