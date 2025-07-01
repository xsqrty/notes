package user

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID             uuid.UUID    `op:"id,primary"`
	Name           string       `op:"name"`
	Email          string       `op:"email"`
	HashedPassword string       `op:"hashed_password"`
	CreatedAt      time.Time    `op:"created_at"`
	UpdatedAt      sql.NullTime `op:"updated_at"`
}
