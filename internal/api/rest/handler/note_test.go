package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/adapter/dtoadapter"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/internal/middleware"
	"github.com/xsqrty/notes/mocks/app/mock_app"
	"github.com/xsqrty/notes/mocks/domain/mock_note"
	"github.com/xsqrty/notes/mocks/middleware/mock_middleware"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

	cases := []struct {
		name        string
		id          string
		statusCode  int
		expected    *dto.NoteResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service)
	}{
		{
			name:       "successful_get",
			id:         id.String(),
			statusCode: http.StatusOK,
			expected:   dtoadapter.NoteToResponseDto(n),
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Get(mock.Anything, u, id).Return(n, nil).Once()
			},
		},
		{
			name:       "user_unauthorized",
			id:         id.String(),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			name:       "param_error",
			id:         "1",
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			name:       "note_not_found",
			id:         id.String(),
			statusCode: http.StatusNotFound,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Get(mock.Anything, u, id).Return(nil, note.ErrNoteNotFound).Once()
			},
		},
		{
			name:       "not_granted",
			id:         id.String(),
			statusCode: http.StatusForbidden,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Get(mock.Anything, u, id).Return(nil, note.ErrNoteOperationForbiddenForUser).Once()
			},
		},
		{
			name:       "unknown_error",
			id:         id.String(),
			statusCode: http.StatusInternalServerError,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Get(mock.Anything, u, id).Return(nil, errors.New("unknown error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.NoteService = service
				deps.JWTAuthentication = mw
			})

			handler := NewNoteHandler(deps)
			if tc.mocker != nil {
				tc.mocker(mw, service)
			}

			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/notes/%s", tc.id), nil)
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.Get)).ServeHTTP(w, addUrlParams(t, r, map[string]string{
				"id": tc.id,
			}))
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.NoteResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

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

	cases := []struct {
		name        string
		req         *dto.NoteRequest
		statusCode  int
		expected    *dto.NoteResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service)
	}{
		{
			name:       "successful_created",
			statusCode: http.StatusCreated,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expected: dtoadapter.NoteToResponseDto(n),
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Create(mock.Anything, u, dtoadapter.NoteRequestDtoToCreateData(req)).Return(n, nil).Once()
			},
		},
		{
			name:       "user_unauthorized",
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			name:       "request_error",
			statusCode: http.StatusBadRequest,
			req:        &dto.NoteRequest{},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			name:       "not_granted",
			statusCode: http.StatusForbidden,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Create(mock.Anything, u, dtoadapter.NoteRequestDtoToCreateData(req)).Return(nil, note.ErrNoteOperationForbiddenForUser).Once()
			},
		},
		{
			name:       "unknown_error",
			statusCode: http.StatusInternalServerError,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Create(mock.Anything, u, dtoadapter.NoteRequestDtoToCreateData(req)).Return(nil, errors.New("unknown error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.NoteService = service
				deps.JWTAuthentication = mw
			})

			handler := NewNoteHandler(deps)
			if tc.mocker != nil {
				tc.mocker(tc.req, mw, service)
			}

			body, err := json.Marshal(tc.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(body))
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.Create)).ServeHTTP(w, r)
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.NoteResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

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

	cases := []struct {
		name        string
		id          string
		req         *dto.NoteRequest
		statusCode  int
		expected    *dto.NoteResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service)
	}{
		{
			name:       "successful_updated",
			id:         id.String(),
			statusCode: http.StatusOK,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expected: dtoadapter.NoteToResponseDto(n),
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).Return(n, nil).Once()
			},
		},
		{
			name:       "user_unauthorized",
			id:         id.String(),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			name:       "request_error",
			id:         id.String(),
			statusCode: http.StatusBadRequest,
			req:        &dto.NoteRequest{},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			name:       "note_not_found",
			id:         id.String(),
			statusCode: http.StatusNotFound,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).Return(nil, note.ErrNoteNotFound).Once()
			},
		},
		{
			name:       "not_granted",
			id:         id.String(),
			statusCode: http.StatusForbidden,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).Return(nil, note.ErrNoteOperationForbiddenForUser).Once()
			},
		},
		{
			name:       "unknown_error",
			id:         id.String(),
			statusCode: http.StatusInternalServerError,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Update(mock.Anything, u, dtoadapter.NoteRequestDtoToUpdateData(id, req)).Return(nil, errors.New("unknown error")).Once()
			},
		},
		{
			name:       "param_error",
			id:         "1",
			statusCode: http.StatusBadRequest,
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			mocker: func(req *dto.NoteRequest, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.NoteService = service
				deps.JWTAuthentication = mw
			})

			handler := NewNoteHandler(deps)
			if tc.mocker != nil {
				tc.mocker(tc.req, mw, service)
			}

			body, err := json.Marshal(tc.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/notes/%s", tc.id), bytes.NewReader(body))
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.Update)).ServeHTTP(w, addUrlParams(t, r, map[string]string{
				"id": tc.id,
			}))
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.NoteResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

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

	cases := []struct {
		name        string
		id          string
		statusCode  int
		expected    *dto.NoteResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service)
	}{
		{
			name:       "successful_get",
			id:         id.String(),
			statusCode: http.StatusOK,
			expected:   dtoadapter.NoteToResponseDto(n),
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Delete(mock.Anything, u, id).Return(n, nil).Once()
			},
		},
		{
			name:       "user_unauthorized",
			id:         id.String(),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			name:       "param_error",
			id:         "1",
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			name:       "note_not_found",
			id:         id.String(),
			statusCode: http.StatusNotFound,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Delete(mock.Anything, u, id).Return(nil, note.ErrNoteNotFound).Once()
			},
		},
		{
			name:       "not_granted",
			id:         id.String(),
			statusCode: http.StatusForbidden,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Delete(mock.Anything, u, id).Return(nil, note.ErrNoteOperationForbiddenForUser).Once()
			},
		},
		{
			name:       "unknown_error",
			id:         id.String(),
			statusCode: http.StatusInternalServerError,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			mocker: func(mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Delete(mock.Anything, u, id).Return(nil, errors.New("unknown error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.NoteService = service
				deps.JWTAuthentication = mw
			})

			handler := NewNoteHandler(deps)
			if tc.mocker != nil {
				tc.mocker(mw, service)
			}

			r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/notes/%s", tc.id), nil)
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.Delete)).ServeHTTP(w, addUrlParams(t, r, map[string]string{
				"id": tc.id,
			}))
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.NoteResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

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

	cases := []struct {
		name        string
		req         any
		statusCode  int
		expected    *dto.NoteSearchResponse
		expectedErr *httpio.ErrorResponse
		mocker      func(req any, mw *mock_middleware.JWTAuthentication, service *mock_note.Service)
	}{
		{
			name:       "successful_search",
			statusCode: http.StatusOK,
			req:        &search.Request{},
			expected:   dtoadapter.NoteSearchToResponseDto(searchResult),
			mocker: func(req any, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Search(mock.Anything, u, req).Return(searchResult, nil).Once()
			},
		},
		{
			name:       "request_error",
			statusCode: http.StatusBadRequest,
			req: map[string]string{
				"filters": "test",
			},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeJsonParse,
				},
			},
			mocker: func(req any, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
			},
		},
		{
			name:       "user_unauthorized",
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
			mocker: func(req any, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(nil, errors.New("no user")).Once()
			},
		},
		{
			name:       "invalid_filter",
			statusCode: http.StatusBadRequest,
			req:        &search.Request{},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
			mocker: func(req any, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Search(mock.Anything, u, req).Return(nil, note.ErrNoteSearchBadRequest).Once()
			},
		},
		{
			name:       "not_granted",
			statusCode: http.StatusForbidden,
			req:        &search.Request{},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeForbidden,
				},
			},
			mocker: func(req any, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Search(mock.Anything, u, req).Return(nil, note.ErrNoteOperationForbiddenForUser).Once()
			},
		},
		{
			name:       "unknown_error",
			statusCode: http.StatusInternalServerError,
			req:        &search.Request{},
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnknown,
				},
			},
			mocker: func(req any, mw *mock_middleware.JWTAuthentication, service *mock_note.Service) {
				mw.EXPECT().GetUser(mock.Anything).Return(u, nil).Once()
				service.EXPECT().Search(mock.Anything, u, req).Return(nil, errors.New("unknown error")).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := mock_note.NewService(t)
			mw := mock_middleware.NewJWTAuthentication(t)

			deps := mock_app.NewDeps(t, func(deps *app.Deps) {
				deps.Service.NoteService = service
				deps.JWTAuthentication = mw
			})

			handler := NewNoteHandler(deps)
			if tc.mocker != nil {
				tc.mocker(tc.req, mw, service)
			}

			body, err := json.Marshal(tc.req)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/notes/search", bytes.NewReader(body))
			w := httptest.NewRecorder()

			middleware.Logger(deps.Logger)(http.HandlerFunc(handler.Search)).ServeHTTP(w, r)
			res := w.Result()

			require.Equal(t, tc.statusCode, res.StatusCode)
			if tc.expected != nil {
				result := dto.NoteSearchResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
				require.Equal(t, tc.expected, &result)
			} else {
				err := httpio.ErrorResponse{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
				require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
				require.NotEmpty(t, err.Error.Message)
			}

			mock.AssertExpectationsForObjects(t, service, mw)
		})
	}
}

func addUrlParams(t *testing.T, r *http.Request, params map[string]string) *http.Request {
	t.Helper()
	ctx := chi.NewRouteContext()
	for k, v := range params {
		ctx.URLParams.Add(k, v)
	}

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}
