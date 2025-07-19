package role

import (
	"context"

	"github.com/xsqrty/notes/internal/domain/user"
)

// Repository defines methods for managing user roles and permissions within the system.
type Repository interface {
	AttachUserRolesByLabel(ctx context.Context, label Label, user *user.User) error
	HasPermissions(ctx context.Context, permissions []Permission, user *user.User) (bool, error)
}
