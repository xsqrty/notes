package role

import (
	"github.com/google/uuid"
	"github.com/xsqrty/op/driver"
	"time"
)

type Label string
type Permission string

const (
	LabelOnCreated Label = "on_created"
)

type Role struct {
	ID          uuid.UUID       `op:"id,primary"`
	Description string          `op:"description"`
	Permissions []string        `op:"permissions"`
	Label       string          `op:"label"`
	CreatedAt   time.Time       `op:"created_at"`
	UpdatedAt   driver.ZeroTime `op:"updated_at"`
}

type UserRelation struct {
	ID        uuid.UUID `op:"id,primary"`
	RoleID    uuid.UUID `op:"role_id"`
	UserID    uuid.UUID `op:"user_id"`
	CreatedAt time.Time `op:"created_at"`
}
