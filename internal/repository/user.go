package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/op"
	"github.com/xsqrty/op/db"
	"github.com/xsqrty/op/orm"
)

type userRepo struct {
	qe db.ConnPool
}

const usersTableName = "users"

func NewUserRepo(qe db.ConnPool) user.Repository {
	return &userRepo{qe}
}

func (r *userRepo) Save(ctx context.Context, u *user.User) error {
	if u.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			fmt.Errorf("save user (generate uuid): %w", err)
		}

		u.ID = id
	}

	err := orm.Put(usersTableName, u).With(ctx, r.qe)
	if err != nil {
		return fmt.Errorf("save user (put): %w", err)
	}

	return nil
}

func (r *userRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	count, err := orm.Count(usersTableName).By("id").Where(op.Eq("email", email)).With(ctx, r.qe)
	if err != nil {
		return false, fmt.Errorf("check user email error: %w", err)
	}

	return count > 0, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	user, err := orm.Query[user.User](
		op.Select().From(usersTableName).Where(op.Eq("email", email)),
	).GetOne(ctx, r.qe)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	user, err := orm.Query[user.User](
		op.Select().From(usersTableName).Where(op.Eq("id", id)),
	).GetOne(ctx, r.qe)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}
