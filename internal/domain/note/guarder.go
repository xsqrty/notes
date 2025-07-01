package note

import (
	"context"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/rbac"
)

type Guarder interface {
	IsGranted(ctx context.Context, op rbac.Operation, note *Note, user *user.User) (bool, error)
}
