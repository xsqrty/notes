package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("user not found")

// User represents a system user with account-related information.
type User struct {
	ID             uuid.UUID    `op:"id,primary"`
	Name           string       `op:"name"`
	Email          string       `op:"email"`
	HashedPassword string       `op:"hashed_password"`
	CreatedAt      time.Time    `op:"created_at"`
	UpdatedAt      sql.NullTime `op:"updated_at"`
}
