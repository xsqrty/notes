package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines methods for managing User entities in the system.
type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	Save(ctx context.Context, u *User) error
}
