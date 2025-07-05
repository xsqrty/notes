package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"net/http"
	"strconv"
	"testing"
)

func TestIntegrationNote_Get(t *testing.T) {
	t.Parallel()

	token := generateAccessToken(t)
	readyNote := createNote(t, token, &dto.NoteRequest{
		Name: gofakeit.Name(),
		Text: gofakeit.Sentence(5),
	})

	cases := []integrationCase[any, dto.NoteResponse]{
		{
			name:       "successful_get",
			additional: readyNote.ID,
			token:      token,
			statusCode: http.StatusOK,
			expected: &dto.NoteResponse{
				ID:   readyNote.ID,
				Name: readyNote.Name,
				Text: readyNote.Text,
			},
		},
		{
			name:       "bad_id",
			additional: strconv.Itoa(gofakeit.Int()),
			token:      token,
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			name:       "not_found",
			additional: uuid.Must(uuid.NewV7()),
			token:      token,
			statusCode: http.StatusNotFound,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
		},
		{
			name:       "unauthorized",
			additional: uuid.Must(uuid.NewV7()),
			token:      gofakeit.LetterN(20),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodGet, fmt.Sprintf("/api/v1/notes/%s", tc.additional), func(expected, actual *dto.NoteResponse) {
				require.Equal(t, expected.Name, actual.Name)
				require.Equal(t, expected.Text, actual.Text)
				require.NotEmpty(t, actual.CreatedAt)
				require.Condition(t, func() bool {
					return expected.ID == actual.ID
				})
			})
		})
	}
}

func TestIntegrationNote_Create(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(5)

	cases := []integrationCase[dto.NoteRequest, dto.NoteResponse]{
		{
			name: "successful_created",
			tokenFactory: func() string {
				return generateAccessToken(t)
			},
			req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			statusCode: http.StatusCreated,
			expected: &dto.NoteResponse{
				Name: name,
				Text: text,
			},
		},
		{
			name: "bad_request",
			tokenFactory: func() string {
				return generateAccessToken(t)
			},
			req: &dto.NoteRequest{
				Name: name,
			},
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			name:       "unauthorized",
			token:      gofakeit.LetterN(20),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodPost, "/api/v1/notes", func(expected, actual *dto.NoteResponse) {
				require.Equal(t, expected.Name, actual.Name)
				require.Equal(t, expected.Text, actual.Text)
				require.NotEmpty(t, actual.CreatedAt)
				require.Condition(t, func() bool {
					return actual.ID != uuid.Nil
				})
			})
		})
	}
}

func TestIntegrationNote_Update(t *testing.T) {
	t.Parallel()

	newName := gofakeit.Name()
	newText := gofakeit.Sentence(10)

	token := generateAccessToken(t)
	readyNote := createNote(t, token, &dto.NoteRequest{
		Name: gofakeit.Name(),
		Text: gofakeit.Sentence(5),
	})

	cases := []integrationCase[dto.NoteRequest, dto.NoteResponse]{
		{
			name:       "successful_updated",
			additional: readyNote.ID,
			token:      token,
			req: &dto.NoteRequest{
				Name: newName,
				Text: newText,
			},
			statusCode: http.StatusOK,
			expected: &dto.NoteResponse{
				ID:   readyNote.ID,
				Name: newName,
				Text: newText,
			},
		},
		{
			name:       "bad_id",
			additional: strconv.Itoa(gofakeit.Int()),
			token:      token,
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			name:       "bad_request",
			additional: readyNote.ID,
			token:      token,
			req: &dto.NoteRequest{
				Name: newName,
			},
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			name:       "not_found",
			additional: uuid.Must(uuid.NewV7()),
			token:      token,
			req: &dto.NoteRequest{
				Name: gofakeit.Name(),
				Text: gofakeit.Sentence(5),
			},
			statusCode: http.StatusNotFound,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
		},
		{
			name:       "unauthorized",
			additional: uuid.Must(uuid.NewV7()),
			token:      gofakeit.LetterN(20),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodPut, fmt.Sprintf("/api/v1/notes/%s", tc.additional), func(expected, actual *dto.NoteResponse) {
				require.Equal(t, expected.Name, actual.Name)
				require.Equal(t, expected.Text, actual.Text)
				require.NotEmpty(t, actual.CreatedAt)
				require.Condition(t, func() bool {
					return expected.ID == actual.ID
				})
			})
		})
	}
}

func TestIntegrationNote_Delete(t *testing.T) {
	t.Parallel()

	token := generateAccessToken(t)
	readyNote := createNote(t, token, &dto.NoteRequest{
		Name: gofakeit.Name(),
		Text: gofakeit.Sentence(5),
	})

	cases := []integrationCase[any, dto.NoteResponse]{
		{
			name:       "successful_deleted",
			additional: readyNote.ID,
			token:      token,
			statusCode: http.StatusOK,
			expected: &dto.NoteResponse{
				ID:   readyNote.ID,
				Name: readyNote.Name,
				Text: readyNote.Text,
			},
		},
		{
			name:       "bad_id",
			additional: strconv.Itoa(gofakeit.Int()),
			token:      token,
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			name:       "not_found",
			additional: uuid.Must(uuid.NewV7()),
			token:      token,
			statusCode: http.StatusNotFound,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
		},
		{
			name:       "unauthorized",
			additional: uuid.Must(uuid.NewV7()),
			token:      gofakeit.LetterN(20),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodDelete, fmt.Sprintf("/api/v1/notes/%s", tc.additional), func(expected, actual *dto.NoteResponse) {
				require.Equal(t, expected.Name, actual.Name)
				require.Equal(t, expected.Text, actual.Text)
				require.NotEmpty(t, actual.CreatedAt)
				require.Condition(t, func() bool {
					return expected.ID == actual.ID
				})

				require.False(t, noteExists(t, token, readyNote.ID))
			})
		})
	}
}

func TestIntegrationNote_Search(t *testing.T) {
	t.Parallel()

	token := generateAccessToken(t)
	notes := make([]*dto.NoteResponse, 10)
	for i := range notes {
		notes[i] = createNote(t, token, &dto.NoteRequest{
			Name: gofakeit.Name(),
			Text: gofakeit.Sentence(5),
		})
	}

	limitedNotes := notes[:len(notes)/2-1]

	cases := []integrationCase[search.Request, dto.NoteSearchResponse]{
		{
			name:  "successful_search_all",
			token: token,
			req: &search.Request{
				Orders: []search.Order{
					{Key: "id", Desc: false},
				},
				Limit: uint64(len(notes)),
			},
			statusCode: http.StatusOK,
			expected: &dto.NoteSearchResponse{
				TotalRows: uint64(len(notes)),
				Rows:      notes,
			},
		},
		{
			name:  "successful_search_limited",
			token: token,
			req: &search.Request{
				Orders: []search.Order{
					{Key: "id", Desc: false},
				},
				Limit: uint64(len(limitedNotes)),
			},
			statusCode: http.StatusOK,
			expected: &dto.NoteSearchResponse{
				TotalRows: uint64(len(notes)),
				Rows:      limitedNotes,
			},
		},
		{
			name:  "successful_search_filtered",
			token: token,
			req: &search.Request{
				Orders: []search.Order{
					{Key: "id", Desc: false},
				},
				Limit: uint64(len(notes)),
				Filters: map[string]any{
					"$or": []map[string]any{
						{
							"id": notes[0].ID,
						},
						{
							"id": notes[1].ID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
			expected: &dto.NoteSearchResponse{
				TotalRows: 2,
				Rows:      notes[:2],
			},
		},
		{
			name:  "successful_search_filtered_offset",
			token: token,
			req: &search.Request{
				Orders: []search.Order{
					{Key: "id", Desc: false},
				},
				Limit:  1,
				Offset: 1,
				Filters: map[string]any{
					"$or": []map[string]any{
						{
							"id": notes[0].ID,
						},
						{
							"id": notes[1].ID,
						},
					},
				},
			},
			statusCode: http.StatusOK,
			expected: &dto.NoteSearchResponse{
				TotalRows: 2,
				Rows:      notes[1:2],
			},
		},
		{
			name:  "unavailable_field",
			token: token,
			req: &search.Request{
				Filters: map[string]any{
					"unavailable": "1",
				},
			},
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			name:       "bad_request",
			token:      token,
			statusCode: http.StatusBadRequest,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeJsonParse,
				},
			},
		},
		{
			name:       "unauthorized",
			token:      gofakeit.LetterN(20),
			statusCode: http.StatusUnauthorized,
			expectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t, http.MethodPost, "/api/v1/notes/search", func(expected, actual *dto.NoteSearchResponse) {
				require.Equal(t, tc.expected.TotalRows, actual.TotalRows)
				require.Len(t, actual.Rows, len(tc.expected.Rows))

				for i, expected := range tc.expected.Rows {
					require.Equal(t, expected.Name, actual.Rows[i].Name)
					require.Equal(t, expected.Text, actual.Rows[i].Text)
					require.NotEmpty(t, actual.Rows[i].CreatedAt)
					require.Condition(t, func() bool {
						return expected.ID == actual.Rows[i].ID
					})
				}
			})
		})
	}
}

func createNote(t *testing.T, token string, req *dto.NoteRequest) *dto.NoteResponse {
	t.Helper()
	jsonReq, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodPost, withBaseUrl("/api/v1/notes"), bytes.NewBuffer(jsonReq))
	require.NoError(t, err)
	httpReq.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	n := &dto.NoteResponse{}
	json.NewDecoder(res.Body).Decode(n)
	require.Equal(t, req.Name, n.Name)
	require.Equal(t, req.Text, n.Text)
	require.NotEmpty(t, n.CreatedAt)
	require.Condition(t, func() bool {
		return n.ID != uuid.Nil
	})

	return n
}

func noteExists(t *testing.T, token string, id uuid.UUID) bool {
	t.Helper()

	httpReq, err := http.NewRequest(http.MethodGet, withBaseUrl(fmt.Sprintf("/api/v1/notes/%s", id)), nil)
	require.NoError(t, err)
	httpReq.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Condition(t, func() bool {
		return res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusOK
	})

	return http.StatusNotFound != res.StatusCode
}
