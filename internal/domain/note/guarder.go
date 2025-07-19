package note

import (
	"context"

	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/rbac"
)

// Guarder defines an interface for determining if a user has permission to perform an operation on a resource.
type Guarder interface {
	IsGranted(ctx context.Context, op rbac.Operation, note *Note, user *user.User) (bool, error)
}
