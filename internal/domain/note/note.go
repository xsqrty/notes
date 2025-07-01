package note

import (
	"errors"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/op/driver"
	"time"
)

var (
	ErrNoteNotFound                  = errors.New("note not found")
	ErrNoteSearchBadRequest          = errors.New("search bad request")
	ErrNoteOperationForbiddenForUser = errors.New("note operation is forbidden for user")
)

const (
	PermissionCreate role.Permission = "notes.create"
	PermissionDelete role.Permission = "notes.delete"
	PermissionUpdate role.Permission = "notes.update"
	PermissionRead   role.Permission = "notes.read"
)

type Note struct {
	ID        uuid.UUID       `op:"id,primary"`
	Name      string          `op:"name"`
	Text      string          `op:"text"`
	UserId    uuid.UUID       `op:"user_id"`
	CreatedAt time.Time       `op:"created_at"`
	UpdatedAt driver.ZeroTime `op:"updated_at"`
}

type UpdateData struct {
	ID   uuid.UUID
	Name string
	Text string
}

type CreateData struct {
	Name string
	Text string
}
