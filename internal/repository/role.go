package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/op"
	"github.com/xsqrty/op/db"
	"github.com/xsqrty/op/orm"
	"time"
)

type roleRepo struct {
	qe db.ConnPool
}

const (
	rolesTableName      = "roles"
	rolesUsersTableName = "roles_users"
)

func NewRoleRepository(qe db.ConnPool) role.Repository {
	return &roleRepo{qe: qe}
}

func (rr *roleRepo) AttachUserRolesByLabel(ctx context.Context, label role.Label, u *user.User) error {
	roles, err := orm.Query[role.Role](op.Select("id").From(rolesTableName).Where(op.Eq("label", label))).GetMany(ctx, rr.qe)
	if err != nil {
		return fmt.Errorf("attach roles with label for user (query roles) %w (label %s, user %s)", err, label, u.ID)
	}

	for _, r := range roles {
		count, err := orm.Count(rolesUsersTableName).Where(op.And{op.Eq("role_id", r.ID), op.Eq("user_id", u.ID)}).With(ctx, rr.qe)
		if err != nil {
			return fmt.Errorf("attach roles with label for user (count roles) %w (label %s, user %s, role %s)", err, label, u.ID, r.ID)
		}

		if count > 0 {
			continue
		}

		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("attach roles with label for user (uuid generate) %w (label %s, user %s, role %s)", err, label, u.ID, r.ID)
		}

		err = orm.Put(rolesUsersTableName, &role.UserRelation{
			ID:        id,
			UserID:    u.ID,
			RoleID:    r.ID,
			CreatedAt: time.Now(),
		}).With(ctx, rr.qe)

		if err != nil {
			return fmt.Errorf("attach roles with label for user (put role) %w (label %s, user %s, role %s)", err, label, u.ID, r.ID)
		}
	}

	return nil
}

func (rr *roleRepo) HasPermissions(ctx context.Context, permissions []role.Permission, u *user.User) (bool, error) {
	count, err := orm.Count(rolesUsersTableName).
		Join(rolesTableName, op.Eq("role_id", op.Column("roles.id"))).
		Where(op.And{
			op.Eq("user_id", u.ID),
			op.Lc("permissions", permissions),
		}).With(ctx, rr.qe)
	if err != nil {
		return false, fmt.Errorf("check roles permissions %w (user %s, permissions %v)", err, u.ID, permissions)
	}

	return count > 0, nil
}
