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

// userRepo is a concrete implementation of the user.Repository interface for managing user entities within a database.
// It uses a db.ConnPool for database interaction and query execution.
type userRepo struct {
	qe db.ConnPool
}

// usersTableName specifies the name of the database table used for storing user records.
const usersTableName = "users"

// NewUserRepo initializes and returns a user.Repository implementation using the provided database connection pool.
func NewUserRepo(qe db.ConnPool) user.Repository {
	return &userRepo{qe}
}

// Save persists a user to the database. Generates a new UUID for the created user.
func (r *userRepo) Save(ctx context.Context, u *user.User) error {
	if u.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("save user (generate uuid): %w", err)
		}

		u.ID = id
	}

	err := orm.Put(usersTableName, u).With(ctx, r.qe)
	if err != nil {
		return fmt.Errorf("save user (put): %w", err)
	}

	return nil
}

// EmailExists checks if an email exists in the user's table. It returns true if the email is found, otherwise false.
func (r *userRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	count, err := orm.Count(op.Select().From(usersTableName).Where(op.Eq("email", email))).By("id").With(ctx, r.qe)
	if err != nil {
		return false, fmt.Errorf("check user email error: %w", err)
	}

	return count > 0, nil
}

// GetByEmail retrieves a user by their email address from the table.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	user, err := orm.Query[user.User](
		op.Select().From(usersTableName).Where(op.Eq("email", email)),
	).GetOne(ctx, r.qe)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by their ID from the database.
func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	user, err := orm.Query[user.User](
		op.Select().From(usersTableName).Where(op.Eq("id", id)),
	).GetOne(ctx, r.qe)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}
