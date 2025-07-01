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

type noteRepo struct {
	qe db.ConnPool
}

const notesTableName = "notes"

func NewNoteRepo(qe db.ConnPool) note.Repository {
	return &noteRepo{qe}
}

func (r *noteRepo) Save(ctx context.Context, n *note.Note) error {
	if n.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			fmt.Errorf("save note (generate uuid): %w", err)
		}

		n.ID = id
	}

	err := orm.Put(notesTableName, n).With(ctx, r.qe)
	if err != nil {
		return fmt.Errorf("save note: %w", err)
	}

	return nil
}

func (r *noteRepo) GetByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	user, err := orm.Query[note.Note](
		op.Select().From(notesTableName).Where(op.Eq("id", id)),
	).GetOne(ctx, r.qe)
	if err != nil {
		return nil, fmt.Errorf("get note by id: %w", err)
	}

	return user, nil
}

func (r *noteRepo) Delete(ctx context.Context, n *note.Note) error {
	_, err := orm.Exec(
		op.Delete(notesTableName).Where(op.Eq("id", n.ID)),
	).With(ctx, r.qe)

	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (r *noteRepo) IDExists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := orm.Count(notesTableName).Where(op.Eq("id", id)).With(ctx, r.qe)
	if err != nil {
		return false, fmt.Errorf("check note id: %w", err)
	}

	return count > 0, nil
}

func (r *noteRepo) SearchByUser(ctx context.Context, u *user.User, req *search.Request) (*search.Result[note.Note], error) {
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
