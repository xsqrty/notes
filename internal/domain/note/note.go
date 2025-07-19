package note

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/op/driver"
)

var (
	ErrNoteNotFound                  = errors.New("note not found")
	ErrNoteSearchBadRequest          = errors.New("search bad request")
	ErrNoteOperationForbiddenForUser = errors.New("note operation is forbidden for user")
)

const (
	// PermissionCreate grants the ability to create notes.
	PermissionCreate role.Permission = "notes.create"
	// PermissionDelete grants the ability to delete notes.
	PermissionDelete role.Permission = "notes.delete"
	// PermissionUpdate grants the ability to update notes.
	PermissionUpdate role.Permission = "notes.update"
	// PermissionRead grants the ability to read notes.
	PermissionRead role.Permission = "notes.read"
)

// Note structure
type Note struct {
	ID        uuid.UUID       `op:"id,primary"`
	Name      string          `op:"name"`
	Text      string          `op:"text"`
	UserId    uuid.UUID       `op:"user_id"`
	CreatedAt time.Time       `op:"created_at"`
	UpdatedAt driver.ZeroTime `op:"updated_at"`
}

// UpdateData represents the data required to update an existing note.
type UpdateData struct {
	ID   uuid.UUID
	Name string
	Text string
}

// CreateData represents the data required to create a new note.
type CreateData struct {
	Name string
	Text string
}
