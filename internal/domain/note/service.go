package note

import (
	"context"

	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
)

// Service notes service interface
type Service interface {
	Get(ctx context.Context, user *user.User, id uuid.UUID) (*Note, error)
	Create(ctx context.Context, user *user.User, data *CreateData) (*Note, error)
	Update(ctx context.Context, user *user.User, data *UpdateData) (*Note, error)
	Delete(ctx context.Context, user *user.User, id uuid.UUID) (*Note, error)
	Search(ctx context.Context, user *user.User, req *search.Request) (*search.Result[Note], error)
}
