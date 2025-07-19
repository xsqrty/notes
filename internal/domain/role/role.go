package role

import (
	"time"

	"github.com/google/uuid"
	"github.com/xsqrty/op/driver"
)

type (
	// Label represents a string type that can be used for labeling or naming purposes.
	Label string
	// Permission represents a string type that defines access control or authorization levels.
	Permission string
)

const (
	// LabelOnCreated represents a predefined label for actions triggered when a resource is created.
	LabelOnCreated Label = "on_created"
)

// Role represents a user role in the system, containing metadata and associated permissions.
type Role struct {
	ID          uuid.UUID       `op:"id,primary"`
	Description string          `op:"description"`
	Permissions []string        `op:"permissions"`
	Label       string          `op:"label"`
	CreatedAt   time.Time       `op:"created_at"`
	UpdatedAt   driver.ZeroTime `op:"updated_at"`
}

// UserRelation represents a connection between a user and a role, including metadata.
type UserRelation struct {
	ID        uuid.UUID `op:"id,primary"`
	RoleID    uuid.UUID `op:"role_id"`
	UserID    uuid.UUID `op:"user_id"`
	CreatedAt time.Time `op:"created_at"`
}
