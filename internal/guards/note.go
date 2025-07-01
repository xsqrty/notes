package guards

import (
	"context"
	"fmt"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/rbac"
)

func NewNoteGuarder(roleRepo role.Repository) note.Guarder {
	return rbac.NewRBAC[*note.Note, *user.User](func(ctx context.Context, operation rbac.Operation, n *note.Note, u *user.User) (bool, error) {
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
	})
}

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

func isNoteDeleteGranted(ctx context.Context, roleRepo role.Repository, n *note.Note, u *user.User) (bool, error) {
	has, err := roleRepo.HasPermissions(ctx, []role.Permission{note.PermissionDelete}, u)
	if !has {
		return false, err
	}

	return n.UserId == u.ID, nil
}

func isNoteUpdateGranted(ctx context.Context, roleRepo role.Repository, n *note.Note, u *user.User) (bool, error) {
	has, err := roleRepo.HasPermissions(ctx, []role.Permission{note.PermissionUpdate}, u)
	if !has {
		return false, err
	}

	return n.UserId == u.ID, nil
}

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
