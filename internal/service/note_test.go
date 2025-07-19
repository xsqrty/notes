package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/mocks/domain/mock_note"
	"github.com/xsqrty/notes/pkg/rbac"
)

func TestNoteService_Create(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(10)
	userId := uuid.Must(uuid.NewV7())
	u := &user.User{
		ID: userId,
	}

	cases := []struct {
		name        string
		user        *user.User
		expected    *note.Note
		expectedErr string
		mocker      func(repo *mock_note.Repository, guard *mock_note.Guarder)
	}{
		{
			name: "successful_create",
			user: u,
			expected: &note.Note{
				Name: name,
				Text: text,
			},
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().IsGranted(mock.Anything, rbac.CREATE, (*note.Note)(nil), u).Return(true, nil).Once()
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:        "save_err",
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("create note: save err (user %s)", u.ID),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().IsGranted(mock.Anything, rbac.CREATE, (*note.Note)(nil), u).Return(true, nil).Once()
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errors.New("save err")).Once()
			},
		},
		{
			name:        "not_granted",
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("create note: note operation is forbidden for user (user %s)", u.ID),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().IsGranted(mock.Anything, rbac.CREATE, (*note.Note)(nil), u).Return(false, nil).Once()
			},
		},
		{
			name:        "granted_error",
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("create note: check granted: granted err (user %s)", u.ID),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().
					IsGranted(mock.Anything, rbac.CREATE, (*note.Note)(nil), u).
					Return(false, errors.New("granted err")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guard := mock_note.NewGuarder(t)
			repo := mock_note.NewRepository(t)
			tc.mocker(repo, guard)

			service := NewNoteService(&NoteServiceDeps{NoteRepo: repo, NoteGuard: guard})
			result, err := service.Create(context.Background(), tc.user, &note.CreateData{
				Name: name,
				Text: text,
			})

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			if tc.expected != nil {
				require.Equal(t, tc.expected.Name, result.Name)
				require.Equal(t, tc.expected.Text, result.Text)
				require.Equal(t, tc.user.ID, result.UserId)
				require.NotZero(t, result.CreatedAt)
			}

			mock.AssertExpectationsForObjects(t, repo, guard)
		})
	}
}

func TestNoteService_Get(t *testing.T) {
	t.Parallel()

	id := uuid.Must(uuid.NewV7())
	userId := uuid.Must(uuid.NewV7())
	n := &note.Note{
		ID:     id,
		Name:   gofakeit.Name(),
		Text:   gofakeit.Sentence(10),
		UserId: userId,
	}

	u := &user.User{
		ID: userId,
	}

	cases := []struct {
		name        string
		id          uuid.UUID
		user        *user.User
		expected    *note.Note
		expectedErr string
		mocker      func(repo *mock_note.Repository, guard *mock_note.Guarder)
	}{
		{
			name:     "successful_get",
			id:       id,
			user:     u,
			expected: n,
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().IsGranted(mock.Anything, rbac.READ, n, u).Return(true, nil).Once()
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
			},
		},
		{
			name:        "note_not_found",
			id:          id,
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("get note: note not found\nno rows (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(nil, errors.New("no rows")).Once()
			},
		},
		{
			name:        "not_granted",
			id:          id,
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("get note: note operation is forbidden for user (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().IsGranted(mock.Anything, rbac.READ, n, u).Return(false, nil).Once()
			},
		},
		{
			name:        "granted_error",
			id:          id,
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("get note: check granted: granted error (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().
					IsGranted(mock.Anything, rbac.READ, n, u).
					Return(false, errors.New("granted error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guard := mock_note.NewGuarder(t)
			repo := mock_note.NewRepository(t)
			tc.mocker(repo, guard)

			service := NewNoteService(&NoteServiceDeps{NoteRepo: repo, NoteGuard: guard})
			result, err := service.Get(context.Background(), tc.user, tc.id)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			require.Equal(t, tc.expected, result)
			mock.AssertExpectationsForObjects(t, repo, guard)
		})
	}
}

func TestNoteService_Update(t *testing.T) {
	t.Parallel()

	id := uuid.Must(uuid.NewV7())
	name := gofakeit.Name()
	text := gofakeit.Sentence(10)
	userId := uuid.Must(uuid.NewV7())

	createNote := func() *note.Note {
		return &note.Note{
			ID:     id,
			Name:   name,
			Text:   text,
			UserId: userId,
		}
	}

	u := &user.User{
		ID: userId,
	}

	cases := []struct {
		name        string
		user        *user.User
		expected    *note.Note
		expectedErr string
		mocker      func(repo *mock_note.Repository, guard *mock_note.Guarder)
	}{
		{
			name:     "successful_update",
			user:     u,
			expected: createNote(),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				n := createNote()
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().IsGranted(mock.Anything, rbac.UPDATE, n, u).Return(true, nil).Once()
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:        "note_not_found",
			user:        u,
			expectedErr: fmt.Sprintf("update note: note not found\nno rows (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(nil, errors.New("no rows")).Once()
			},
		},
		{
			name:        "not_granted",
			user:        u,
			expectedErr: fmt.Sprintf("update note: note operation is forbidden for user (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				n := createNote()
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().IsGranted(mock.Anything, rbac.UPDATE, n, u).Return(false, nil).Once()
			},
		},
		{
			name:        "granted_error",
			user:        u,
			expectedErr: fmt.Sprintf("update note: check granted: granted error (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				n := createNote()
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().
					IsGranted(mock.Anything, rbac.UPDATE, n, u).
					Return(false, errors.New("granted error")).
					Once()
			},
		},
		{
			name:        "save_error",
			user:        u,
			expectedErr: fmt.Sprintf("update note: connection unavailable (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				n := createNote()
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().IsGranted(mock.Anything, rbac.UPDATE, n, u).Return(true, nil).Once()
				repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errors.New("connection unavailable")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guard := mock_note.NewGuarder(t)
			repo := mock_note.NewRepository(t)
			tc.mocker(repo, guard)

			service := NewNoteService(&NoteServiceDeps{NoteRepo: repo, NoteGuard: guard})
			result, err := service.Update(context.Background(), tc.user, &note.UpdateData{
				ID:   id,
				Name: name,
				Text: text,
			})

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			if tc.expected != nil {
				require.Equal(t, tc.expected.Name, result.Name)
				require.Equal(t, tc.expected.Text, result.Text)
				require.Equal(t, tc.user.ID, result.UserId)
				require.NotZero(t, result.UpdatedAt)
			}

			mock.AssertExpectationsForObjects(t, repo, guard)
		})
	}
}

func TestNoteService_Delete(t *testing.T) {
	t.Parallel()

	id := uuid.Must(uuid.NewV7())
	userId := uuid.Must(uuid.NewV7())
	n := &note.Note{
		ID:     id,
		Name:   gofakeit.Name(),
		Text:   gofakeit.Sentence(10),
		UserId: userId,
	}

	u := &user.User{
		ID: userId,
	}

	cases := []struct {
		name        string
		id          uuid.UUID
		user        *user.User
		expected    *note.Note
		expectedErr string
		mocker      func(repo *mock_note.Repository, guard *mock_note.Guarder)
	}{
		{
			name:     "successful_delete",
			id:       id,
			user:     u,
			expected: n,
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().IsGranted(mock.Anything, rbac.DELETE, n, u).Return(true, nil).Once()
				repo.EXPECT().Delete(mock.Anything, n).Return(nil).Once()
			},
		},
		{
			name:        "delete_error",
			id:          id,
			user:        u,
			expectedErr: fmt.Sprintf("delete note: can`t delete (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().IsGranted(mock.Anything, rbac.DELETE, n, u).Return(true, nil).Once()
				repo.EXPECT().Delete(mock.Anything, n).Return(errors.New("can`t delete")).Once()
			},
		},
		{
			name:        "note_not_found",
			id:          id,
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("delete note: note not found\nno rows (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(nil, errors.New("no rows")).Once()
			},
		},
		{
			name:        "not_granted",
			id:          id,
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("delete note: note operation is forbidden for user (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().IsGranted(mock.Anything, rbac.DELETE, n, u).Return(false, nil).Once()
			},
		},
		{
			name:        "granted_error",
			id:          id,
			user:        u,
			expected:    nil,
			expectedErr: fmt.Sprintf("delete note: check granted: granted error (user %s, note %s)", u.ID, id),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				repo.EXPECT().GetByID(mock.Anything, id).Return(n, nil).Once()
				guard.EXPECT().
					IsGranted(mock.Anything, rbac.DELETE, n, u).
					Return(false, errors.New("granted error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guard := mock_note.NewGuarder(t)
			repo := mock_note.NewRepository(t)
			tc.mocker(repo, guard)

			service := NewNoteService(&NoteServiceDeps{NoteRepo: repo, NoteGuard: guard})
			result, err := service.Delete(context.Background(), tc.user, tc.id)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			require.Equal(t, tc.expected, result)
			mock.AssertExpectationsForObjects(t, repo, guard)
		})
	}
}

func TestNoteService_Search(t *testing.T) {
	t.Parallel()

	u := &user.User{
		ID: uuid.Must(uuid.NewV7()),
	}

	req := &search.Request{
		Orders: []search.Order{{Key: "id", Desc: true}},
	}

	result := &search.Result[note.Note]{
		TotalRows: 1,
		Rows: []*note.Note{
			{
				Name: gofakeit.Name(),
				Text: gofakeit.Sentence(10),
			},
		},
	}

	cases := []struct {
		name        string
		user        *user.User
		req         *search.Request
		expected    *search.Result[note.Note]
		expectedErr string
		mocker      func(repo *mock_note.Repository, guard *mock_note.Guarder)
	}{
		{
			name:     "successful_search",
			user:     u,
			req:      req,
			expected: result,
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().IsGranted(mock.Anything, rbac.READ, (*note.Note)(nil), u).Return(true, nil).Once()
				repo.EXPECT().SearchByUser(mock.Anything, u, req).Return(result, nil).Once()
			},
		},
		{
			name:        "search_error",
			user:        u,
			req:         req,
			expectedErr: fmt.Sprintf("search note: db unavailable (user %s)", u.ID),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().IsGranted(mock.Anything, rbac.READ, (*note.Note)(nil), u).Return(true, nil).Once()
				repo.EXPECT().SearchByUser(mock.Anything, u, req).Return(nil, errors.New("db unavailable")).Once()
			},
		},
		{
			name:        "not_granted",
			user:        u,
			req:         req,
			expectedErr: fmt.Sprintf("search note: note operation is forbidden for user (user %s)", u.ID),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().IsGranted(mock.Anything, rbac.READ, (*note.Note)(nil), u).Return(false, nil).Once()
			},
		},
		{
			name:        "granted_error",
			user:        u,
			req:         req,
			expectedErr: fmt.Sprintf("search note: check granted: granted error (user %s)", u.ID),
			mocker: func(repo *mock_note.Repository, guard *mock_note.Guarder) {
				guard.EXPECT().
					IsGranted(mock.Anything, rbac.READ, (*note.Note)(nil), u).
					Return(false, errors.New("granted error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			guard := mock_note.NewGuarder(t)
			repo := mock_note.NewRepository(t)
			tc.mocker(repo, guard)

			service := NewNoteService(&NoteServiceDeps{NoteRepo: repo, NoteGuard: guard})
			result, err := service.Search(context.Background(), tc.user, tc.req)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			}

			if tc.expected != nil {
				require.Equal(t, tc.expected, result)
			}

			mock.AssertExpectationsForObjects(t, repo, guard)
		})
	}
}
