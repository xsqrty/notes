package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/rbac"
	"github.com/xsqrty/op/driver"
)

// NoteServiceDeps represents the dependencies required to construct a note service.
type NoteServiceDeps struct {
	NoteRepo  note.Repository
	NoteGuard note.Guarder
}

// noteService is a struct that implements the note.Service interface for managing notes.
type noteService struct {
	noteRepo note.Repository
	guard    note.Guarder
}

// NewNoteService initializes and returns a new implementation of the note.Service interface using the provided dependencies.
func NewNoteService(deps *NoteServiceDeps) note.Service {
	return &noteService{
		noteRepo: deps.NoteRepo,
		guard:    deps.NoteGuard,
	}
}

// Create generates a new note using the provided data for a user, ensuring that the user has the required permissions.
func (s *noteService) Create(ctx context.Context, u *user.User, data *note.CreateData) (*note.Note, error) {
	granted, err := s.guard.IsGranted(ctx, rbac.CREATE, nil, u)
	if err != nil {
		return nil, fmt.Errorf("create note: check granted: %w (user %s)", err, u.ID)
	}

	if !granted {
		return nil, fmt.Errorf("create note: %w (user %s)", note.ErrOperationForbiddenForUser, u.ID)
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

// Get retrieves a note by its ID if the user has the required permission to access it.
func (s *noteService) Get(ctx context.Context, u *user.User, id uuid.UUID) (*note.Note, error) {
	curNote, err := s.noteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get note: %w (user %s, note %s)", errors.Join(note.ErrNotFound, err), u.ID, id)
	}

	granted, err := s.guard.IsGranted(ctx, rbac.READ, curNote, u)
	if err != nil {
		return nil, fmt.Errorf("get note: check granted: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	if !granted {
		return nil, fmt.Errorf(
			"get note: %w (user %s, note %s)",
			note.ErrOperationForbiddenForUser,
			u.ID,
			curNote.ID,
		)
	}

	return curNote, nil
}

// Update modifies an existing note with the provided data if the user is authorized and the note exists. Returns the updated note.
func (s *noteService) Update(ctx context.Context, u *user.User, data *note.UpdateData) (*note.Note, error) {
	curNote, err := s.noteRepo.GetByID(ctx, data.ID)
	if err != nil {
		return nil, fmt.Errorf(
			"update note: %w (user %s, note %s)",
			errors.Join(note.ErrNotFound, err),
			u.ID,
			data.ID,
		)
	}

	granted, err := s.guard.IsGranted(ctx, rbac.UPDATE, curNote, u)
	if err != nil {
		return nil, fmt.Errorf("update note: check granted: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	if !granted {
		return nil, fmt.Errorf(
			"update note: %w (user %s, note %s)",
			note.ErrOperationForbiddenForUser,
			u.ID,
			curNote.ID,
		)
	}

	curNote.UpdatedAt = driver.ZeroTime(time.Now())
	curNote.Name = data.Name
	curNote.Text = data.Text

	if err := s.noteRepo.Save(ctx, curNote); err != nil {
		return nil, fmt.Errorf("update note: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	return curNote, nil
}

// Delete removes a note by its ID if the user has the required permissions and returns the deleted note or an error.
func (s *noteService) Delete(ctx context.Context, u *user.User, id uuid.UUID) (*note.Note, error) {
	curNote, err := s.noteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("delete note: %w (user %s, note %s)", errors.Join(note.ErrNotFound, err), u.ID, id)
	}

	granted, err := s.guard.IsGranted(ctx, rbac.DELETE, curNote, u)
	if err != nil {
		return nil, fmt.Errorf("delete note: check granted: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	if !granted {
		return nil, fmt.Errorf(
			"delete note: %w (user %s, note %s)",
			note.ErrOperationForbiddenForUser,
			u.ID,
			curNote.ID,
		)
	}

	if err := s.noteRepo.Delete(ctx, curNote); err != nil {
		return nil, fmt.Errorf("delete note: %w (user %s, note %s)", err, u.ID, curNote.ID)
	}

	return curNote, nil
}

// Search performs a search operation for notes belonging to the specified user based on the given request parameters.
func (s *noteService) Search(
	ctx context.Context,
	u *user.User,
	req *search.Request,
) (*search.Result[note.Note], error) {
	granted, err := s.guard.IsGranted(ctx, rbac.READ, nil, u)
	if err != nil {
		return nil, fmt.Errorf("search note: check granted: %w (user %s)", err, u.ID)
	}

	if !granted {
		return nil, fmt.Errorf("search note: %w (user %s)", note.ErrOperationForbiddenForUser, u.ID)
	}

	res, err := s.noteRepo.SearchByUser(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("search note: %w (user %s)", err, u.ID)
	}

	return res, nil
}
