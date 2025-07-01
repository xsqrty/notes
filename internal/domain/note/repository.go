package note

import (
	"context"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Note, error)
	IDExists(ctx context.Context, id uuid.UUID) (bool, error)
	Save(ctx context.Context, n *Note) error
	Delete(ctx context.Context, n *Note) error
	SearchByUser(ctx context.Context, u *user.User, r *search.Request) (*search.Result[Note], error)
}
