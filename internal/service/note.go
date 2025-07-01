package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/rbac"
	"github.com/xsqrty/op/driver"
	"time"
)

type NoteServiceDeps struct {
	NoteRepo  note.Repository
	NoteGuard note.Guarder
}

type noteService struct {
	noteRepo note.Repository
	guard    note.Guarder
}

func NewNoteService(deps *NoteServiceDeps) note.Service {
	return &noteService{
		noteRepo: deps.NoteRepo,
		guard:    deps.NoteGuard,
	}
}

func (s *noteService) Create(ctx context.Context, u *user.User, data *note.CreateData) (*note.Note, error) {
	granted, err := s.guard.IsGranted(ctx, rbac.CREATE, nil, u)
	if err != nil {
		return nil, fmt.Errorf("create note: check granted: %w (user %s)", err, u.ID)
	}

	if !granted {
		return nil, fmt.Errorf("create note: %w (user %s)", note.ErrNoteOperationForbiddenForUser, u.ID)
	}

	n := &note.Note{
		Name:      data.Name,
		Text:      data.Text,
		UserId:    u.ID,
		CreatedAt: time.Now(),
	}

	if err := s.noteRepo.Save(ctx, n); err != nil {
		return nil, fmt.Errorf("create note: %w (user %s)", err, u.ID)
	}

	return n, nil
}

func (s *noteService) Get(ctx context.Context, u *user.User, id uuid.UUID) (*note.Note, error) {
	curNote, err := s.noteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get note: %w (user %s, note %s)", errors.Join(note.ErrNoteNotFound, err), u.ID, id)
	}

	granted, err := s.guard.IsGranted(ctx, rbac.READ, curNote, u)
	if err != nil {
		return nil, fmt.Errorf("get note: check granted: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	if !granted {
		return nil, fmt.Errorf("get note: %w (user %s, note %s)", note.ErrNoteOperationForbiddenForUser, u.ID, curNote.ID)
	}

	return curNote, nil
}

func (s *noteService) Update(ctx context.Context, u *user.User, data *note.UpdateData) (*note.Note, error) {
	curNote, err := s.noteRepo.GetByID(ctx, data.ID)
	if err != nil {
		return nil, fmt.Errorf("update note: %w (user %s, note %s)", errors.Join(note.ErrNoteNotFound, err), u.ID, data.ID)
	}

	granted, err := s.guard.IsGranted(ctx, rbac.UPDATE, curNote, u)
	if err != nil {
		return nil, fmt.Errorf("update note: check granted: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	if !granted {
		return nil, fmt.Errorf("update note: %w (user %s, note %s)", note.ErrNoteOperationForbiddenForUser, u.ID, curNote.ID)
	}

	curNote.UpdatedAt = driver.ZeroTime(time.Now())
	curNote.Name = data.Name
	curNote.Text = data.Text

	if err := s.noteRepo.Save(ctx, curNote); err != nil {
		return nil, fmt.Errorf("update note: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	return curNote, nil
}

func (s *noteService) Delete(ctx context.Context, u *user.User, id uuid.UUID) (*note.Note, error) {
	curNote, err := s.noteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("delete note: %w (user %s, note %s)", errors.Join(note.ErrNoteNotFound, err), u.ID, id)
	}

	granted, err := s.guard.IsGranted(ctx, rbac.DELETE, curNote, u)
	if err != nil {
		return nil, fmt.Errorf("delete note: check granted: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	if !granted {
		return nil, fmt.Errorf("delete note: %w (user %s, note %s)", note.ErrNoteOperationForbiddenForUser, u.ID, curNote.ID)
	}

	if err := s.noteRepo.Delete(ctx, curNote); err != nil {
		return nil, fmt.Errorf("delete note: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	return curNote, nil
}

func (s *noteService) Search(ctx context.Context, u *user.User, req *search.Request) (*search.Result[note.Note], error) {
	granted, err := s.guard.IsGranted(ctx, rbac.READ, nil, u)
	if err != nil {
		return nil, fmt.Errorf("search note: check granted: %w (user %s)", err, u.ID)
	}

	if !granted {
		return nil, fmt.Errorf("search note: %w (user %s)", note.ErrNoteOperationForbiddenForUser, u.ID)
	}

	res, err := s.noteRepo.SearchByUser(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("search note: %w (user %s)", err, u.ID)
	}

	return res, nil
}
