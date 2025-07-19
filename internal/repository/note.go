package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/adapter/dtoadapter"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/op"
	"github.com/xsqrty/op/db"
	"github.com/xsqrty/op/orm"
)

// noteRepo represents a concrete implementation of the note.Repository interface using a database connection pool.
type noteRepo struct {
	qe db.ConnPool
}

// notesTableName defines the name of the database table used to store note records.
const notesTableName = "notes"

// NewNoteRepo initializes and returns a note.Repository implementation using the provided database connection pool.
func NewNoteRepo(qe db.ConnPool) note.Repository {
	return &noteRepo{qe}
}

// Save stores the given note in the database, generating a new UUID for the created note.
func (r *noteRepo) Save(ctx context.Context, n *note.Note) error {
	if n.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("save note (generate uuid): %w", err)
		}

		n.ID = id
	}

	err := orm.Put(notesTableName, n).With(ctx, r.qe)
	if err != nil {
		return fmt.Errorf("save note: %w", err)
	}

	return nil
}

// GetByID retrieves a note from the database by the identifier. Returns the note or an error if not found.
func (r *noteRepo) GetByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	user, err := orm.Query[note.Note](
		op.Select().From(notesTableName).Where(op.Eq("id", id)),
	).GetOne(ctx, r.qe)
	if err != nil {
		return nil, fmt.Errorf("get note by id: %w", err)
	}

	return user, nil
}

// Delete removes the specified note from the database based on ID.
func (r *noteRepo) Delete(ctx context.Context, n *note.Note) error {
	_, err := orm.Exec(
		op.Delete(notesTableName).Where(op.Eq("id", n.ID)),
	).With(ctx, r.qe)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// IDExists checks if a note with the given ID exists in the database.
// It returns true if the note exists, otherwise false and an error if encountered.
func (r *noteRepo) IDExists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := orm.Count(op.Select().From(notesTableName).Where(op.Eq("id", id))).With(ctx, r.qe)
	if err != nil {
		return false, fmt.Errorf("check note id: %w", err)
	}

	return count > 0, nil
}

// SearchByUser retrieves notes associated with a specific user based on the search request parameters and pagination options.
func (r *noteRepo) SearchByUser(
	ctx context.Context,
	u *user.User,
	req *search.Request,
) (*search.Result[note.Note], error) {
	res, err := orm.Paginate[note.Note](notesTableName, dtoadapter.SearchToPaginateRequest(req)).
		WhiteList("id", "name", "created_at", "updated_at").
		Fields(
			op.As("id", op.Column("notes.id")),
			op.As("name", op.Column("notes.name")),
			op.As("text", op.Column("notes.text")),
			op.As("user_id", op.Column("notes.user_id")),
			op.As("created_at", op.Column("notes.created_at")),
			op.As("updated_at", op.Column("notes.updated_at")),
		).
		Where(op.Eq("user_id", u.ID)).
		With(ctx, r.qe)
	if err != nil {
		if errors.Is(err, orm.ErrFilterInvalid) {
			return nil, fmt.Errorf("search note: invalid filter: %w", errors.Join(note.ErrNoteSearchBadRequest, err))
		}

		if errors.Is(err, orm.ErrDisallowedKey) {
			return nil, fmt.Errorf("search note: disallowed key: %w", errors.Join(note.ErrNoteSearchBadRequest, err))
		}

		return nil, fmt.Errorf("search note by user: %w", err)
	}

	return &search.Result[note.Note]{
		Rows:      res.Rows,
		TotalRows: res.TotalRows,
	}, nil
}
