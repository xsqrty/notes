package mock_tx

import (
	"context"

	"github.com/xsqrty/notes/internal/domain/tx"
)

type mockTx struct{}

func NewMockTxManager() tx.Manager {
	return &mockTx{}
}

func (m *mockTx) Transact(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
