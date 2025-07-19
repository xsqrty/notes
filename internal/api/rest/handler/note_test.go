package handler

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/xsqrty/notes/internal/adapter/dtoadapter"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/mocks/app/mock_app"
	"github.com/xsqrty/notes/mocks/domain/mock_note"
	"github.com/xsqrty/notes/mocks/middleware/mock_middleware"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/tests/testutil"
)

type noteDeps struct {
	mw      *mock_middleware.JWTAuthentication
	service *mock_note.Service
}

func TestNoteHandler_Get(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(5)

	id := uuid.Must(uuid.NewV7())
	n := &note.Note{
		Name: name,
		Text: text,
	}
	u := &user.User{
		ID:    uuid.Must(uuid.NewV7()),
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}

	cases := []testutil.HandlerCase[struct{}, *dto.NoteResponse, *noteDeps]{
		{
			Name:       "successful_get",
			ID:         id.String(),
			StatusCode: http.StatusOK,
			Expected:   dtoadapter.NoteToResponseDto(n),
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Get(mock.Anything, u, id).Return(n, nil).Once()
			},
		},
		{
			Name:       "user_unauthorized",
			ID:         id.String(),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			Name:       "param_error",
			ID:         "1",
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			Name:       "note_not_found",
			ID:         id.String(),
			StatusCode: http.StatusNotFound,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Get(mock.Anything, u, id).Return(nil, note.ErrNoteNotFound).Once()
			},
		},
		{
			Name:       "not_granted",
			ID:         id.String(),
			StatusCode: http.StatusForbidden,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Get(mock.Anything, u, id).Return(nil, note.ErrNoteOperationForbiddenForUser).Once()
			},
		},
		{
			Name:       "unknown_error",
			ID:         id.String(),
			StatusCode: http.StatusInternalServerError,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Get(mock.Anything, u, id).Return(nil, errors.New("unknown error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			tc.Run(t, http.MethodGet, fmt.Sprintf("/api/v1/notes/%s", tc.ID), func() *noteDeps {
				return &noteDeps{
					service: service,
					mw:      mw,
				}
			}, func(d *noteDeps) http.HandlerFunc {
				return NewNoteHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.JWTAuthentication = mw
					deps.Service.NoteService = service
				})).Get
			})

			mock.AssertExpectationsForObjects(t, service, mw)
		})
	}
}

func TestNoteHandler_Create(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(5)

	n := &note.Note{
		Name: name,
		Text: text,
	}
	u := &user.User{
		ID:    uuid.Must(uuid.NewV7()),
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}

	cases := []testutil.HandlerCase[*dto.NoteRequest, *dto.NoteResponse, *noteDeps]{
		{
			Name:       "successful_created",
			StatusCode: http.StatusCreated,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			Expected: dtoadapter.NoteToResponseDto(n),
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Create(mock.Anything, u, dtoadapter.NoteRequestDtoToCreateData(req)).
					Return(n, nil).
					Once()
			},
		},
		{
			Name:       "user_unauthorized",
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			Name:       "request_error",
			StatusCode: http.StatusBadRequest,
			Req:        &dto.NoteRequest{},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			Name:       "not_granted",
			StatusCode: http.StatusForbidden,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Create(mock.Anything, u, dtoadapter.NoteRequestDtoToCreateData(req)).
					Return(nil, note.ErrNoteOperationForbiddenForUser).
					Once()
			},
		},
		{
			Name:       "unknown_error",
			StatusCode: http.StatusInternalServerError,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Create(mock.Anything, u, dtoadapter.NoteRequestDtoToCreateData(req)).
					Return(nil, errors.New("unknown error")).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			tc.Run(t, http.MethodPost, "/api/v1/notes", func() *noteDeps {
				return &noteDeps{
					service: service,
					mw:      mw,
				}
			}, func(d *noteDeps) http.HandlerFunc {
				return NewNoteHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.JWTAuthentication = mw
					deps.Service.NoteService = service
				})).Create
			})

			mock.AssertExpectationsForObjects(t, service, mw)
		})
	}
}

func TestNoteHandler_Update(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(5)

	id := uuid.Must(uuid.NewV7())
	n := &note.Note{
		Name: name,
		Text: text,
	}
	u := &user.User{
		ID:    uuid.Must(uuid.NewV7()),
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}

	cases := []testutil.HandlerCase[*dto.NoteRequest, *dto.NoteResponse, *noteDeps]{
		{
			Name:       "successful_updated",
			ID:         id.String(),
			StatusCode: http.StatusOK,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			Expected: dtoadapter.NoteToResponseDto(n),
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).
					Return(n, nil).
					Once()
			},
		},
		{
			Name:       "user_unauthorized",
			ID:         id.String(),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			Name:       "request_error",
			ID:         id.String(),
			StatusCode: http.StatusBadRequest,
			Req:        &dto.NoteRequest{},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			Name:       "note_not_found",
			ID:         id.String(),
			StatusCode: http.StatusNotFound,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).
					Return(nil, note.ErrNoteNotFound).
					Once()
			},
		},
		{
			Name:       "not_granted",
			ID:         id.String(),
			StatusCode: http.StatusForbidden,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).
					Return(nil, note.ErrNoteOperationForbiddenForUser).
					Once()
			},
		},
		{
			Name:       "unknown_error",
			ID:         id.String(),
			StatusCode: http.StatusInternalServerError,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).
					Return(nil, errors.New("unknown error")).
					Once()
			},
		},
		{
			Name:       "param_error",
			ID:         "1",
			StatusCode: http.StatusBadRequest,
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			Mocker: func(req *dto.NoteRequest, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			tc.Run(t, http.MethodPut, fmt.Sprintf("/api/v1/notes/%s", tc.ID), func() *noteDeps {
				return &noteDeps{
					service: service,
					mw:      mw,
				}
			}, func(d *noteDeps) http.HandlerFunc {
				return NewNoteHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.JWTAuthentication = mw
					deps.Service.NoteService = service
				})).Update
			})

			mock.AssertExpectationsForObjects(t, service, mw)
		})
	}
}

func TestNoteHandler_Delete(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(5)

	id := uuid.Must(uuid.NewV7())
	n := &note.Note{
		Name: name,
		Text: text,
	}
	u := &user.User{
		ID:    uuid.Must(uuid.NewV7()),
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}

	cases := []testutil.HandlerCase[struct{}, *dto.NoteResponse, *noteDeps]{
		{
			Name:       "successful_get",
			ID:         id.String(),
			StatusCode: http.StatusOK,
			Expected:   dtoadapter.NoteToResponseDto(n),
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Delete(mock.Anything, u, id).Return(n, nil).Once()
			},
		},
		{
			Name:       "user_unauthorized",
			ID:         id.String(),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			Name:       "param_error",
			ID:         "1",
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			Name:       "note_not_found",
			ID:         id.String(),
			StatusCode: http.StatusNotFound,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Delete(mock.Anything, u, id).Return(nil, note.ErrNoteNotFound).Once()
			},
		},
		{
			Name:       "not_granted",
			ID:         id.String(),
			StatusCode: http.StatusForbidden,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Delete(mock.Anything, u, id).
					Return(nil, note.ErrNoteOperationForbiddenForUser).
					Once()
			},
		},
		{
			Name:       "unknown_error",
			ID:         id.String(),
			StatusCode: http.StatusInternalServerError,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			Mocker: func(_ struct{}, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Delete(mock.Anything, u, id).Return(nil, errors.New("unknown error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			tc.Run(t, http.MethodDelete, fmt.Sprintf("/api/v1/notes/%s", tc.ID), func() *noteDeps {
				return &noteDeps{
					service: service,
					mw:      mw,
				}
			}, func(d *noteDeps) http.HandlerFunc {
				return NewNoteHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.JWTAuthentication = mw
					deps.Service.NoteService = service
				})).Delete
			})

			mock.AssertExpectationsForObjects(t, service, mw)
		})
	}
}

func TestNoteHandler_Search(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(5)

	u := &user.User{
		ID:    uuid.Must(uuid.NewV7()),
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}

	searchResult := &search.Result[note.Note]{
		TotalRows: 1,
		Rows: []*note.Note{
			{
				Name: name,
				Text: text,
			},
		},
	}

	cases := []testutil.HandlerCase[any, *dto.NoteSearchResponse, *noteDeps]{
		{
			Name:       "successful_search",
			StatusCode: http.StatusOK,
			Req:        &search.Request{},
			Expected:   dtoadapter.NoteSearchToResponseDto(searchResult),
			Mocker: func(req any, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Search(mock.Anything, u, req).Return(searchResult, nil).Once()
			},
		},
		{
			Name:       "request_error",
			StatusCode: http.StatusBadRequest,
			Req: map[string]string{
				"filters": "test",
			},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeJsonParse,
				},
			},
			Mocker: func(req any, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			Name:       "user_unauthorized",
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			Mocker: func(req any, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			Name:       "invalid_filter",
			StatusCode: http.StatusBadRequest,
			Req:        &search.Request{},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			Mocker: func(req any, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Search(mock.Anything, u, req).Return(nil, note.ErrNoteSearchBadRequest).Once()
			},
		},
		{
			Name:       "not_granted",
			StatusCode: http.StatusForbidden,
			Req:        &search.Request{},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			Mocker: func(req any, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().
					Search(mock.Anything, u, req).
					Return(nil, note.ErrNoteOperationForbiddenForUser).
					Once()
			},
		},
		{
			Name:       "unknown_error",
			StatusCode: http.StatusInternalServerError,
			Req:        &search.Request{},
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			Mocker: func(req any, d *noteDeps) {
				d.mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				d.service.EXPECT().Search(mock.Anything, u, req).Return(nil, errors.New("unknown error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			tc.Run(t, http.MethodPost, "/api/v1/notes/search", func() *noteDeps {
				return &noteDeps{
					service: service,
					mw:      mw,
				}
			}, func(d *noteDeps) http.HandlerFunc {
				return NewNoteHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
					deps.JWTAuthentication = mw
					deps.Service.NoteService = service
				})).Search
			})

			mock.AssertExpectationsForObjects(t, service, mw)
		})
	}
}
