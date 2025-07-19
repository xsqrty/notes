package guards

import (
	"context"
	"fmt"

	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/rbac"
)

// NewNoteGuarder creates a note.Guarder instance using RBAC logic to determine user permissions for note operations.
func NewNoteGuarder(roleRepo role.Repository) note.Guarder {
	return rbac.NewRBAC[*note.Note, *user.User](
		func(ctx context.Context, operation rbac.Operation, n *note.Note, u *user.User) (bool, error) {
			switch operation {
			case rbac.READ:
				return isNoteReadGranted(ctx, roleRepo, n, u)
			case rbac.DELETE:
				return isNoteDeleteGranted(ctx, roleRepo, n, u)
			case rbac.UPDATE:
				return isNoteUpdateGranted(ctx, roleRepo, n, u)
			case rbac.CREATE:
				return isNoteCreateGranted(ctx, roleRepo, n, u)
			}
			return false, fmt.Errorf("note operation %q (%d) is not described", operation, operation)
		},
	)
}

// isNoteReadGranted determines if a user has the permission to read a note based on their roles and note ownership.
func isNoteReadGranted(ctx context.Context, roleRepo role.Repository, n *note.Note, u *user.User) (bool, error) {
	has, err := roleRepo.HasPermissions(ctx, []role.Permission{note.PermissionRead}, u)
	if !has {
		return false, err
	}

	if n == nil {
		return true, nil
	}

	return n.UserId == u.ID, nil
}

// isNoteDeleteGranted checks if a user has the required permission and ownership to delete a specific note.
func isNoteDeleteGranted(ctx context.Context, roleRepo role.Repository, n *note.Note, u *user.User) (bool, error) {
	has, err := roleRepo.HasPermissions(ctx, []role.Permission{note.PermissionDelete}, u)
	if !has {
		return false, err
	}

	return n.UserId == u.ID, nil
}

// isNoteUpdateGranted checks if a user is allowed to update a given note based on their permissions and ownership.
func isNoteUpdateGranted(ctx context.Context, roleRepo role.Repository, n *note.Note, u *user.User) (bool, error) {
	has, err := roleRepo.HasPermissions(ctx, []role.Permission{note.PermissionUpdate}, u)
	if !has {
		return false, err
	}

	return n.UserId == u.ID, nil
}

// isNoteCreateGranted determines if a user is authorized to create a note based on roles, permissions, and note ownership.
func isNoteCreateGranted(ctx context.Context, roleRepo role.Repository, n *note.Note, u *user.User) (bool, error) {
	has, err := roleRepo.HasPermissions(ctx, []role.Permission{note.PermissionCreate}, u)
	if !has {
		return false, err
	}

	if n == nil {
		return true, nil
	}

	return n.UserId == u.ID, nil
}
