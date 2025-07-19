package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/tests/testutil"
)

func TestIntegrationNote_Get(t *testing.T) {
	t.Parallel()

	token := generateAccessToken(t)
	readyNote := createNote(t, token, &dto.NoteRequest{
		Name: gofakeit.Name(),
		Text: gofakeit.Sentence(5),
	})

	cases := []testutil.IntegrationCase[any, dto.NoteResponse]{
		{
			Name:       "successful_get",
			Additional: readyNote.ID,
			Token:      token,
			StatusCode: http.StatusOK,
			Expected: &dto.NoteResponse{
				ID:   readyNote.ID,
				Name: readyNote.Name,
				Text: readyNote.Text,
			},
		},
		{
			Name:       "bad_id",
			Additional: strconv.Itoa(gofakeit.Int()),
			Token:      token,
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			Name:       "not_found",
			Additional: uuid.Must(uuid.NewV7()),
			Token:      token,
			StatusCode: http.StatusNotFound,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
		},
		{
			Name:       "unauthorized",
			Additional: uuid.Must(uuid.NewV7()),
			Token:      gofakeit.LetterN(20),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(
				t,
				http.MethodGet,
				fmt.Sprintf("/api/v1/notes/%s", tc.Additional),
				func(expected, actual *dto.NoteResponse) {
					require.Equal(t, expected.Name, actual.Name)
					require.Equal(t, expected.Text, actual.Text)
					require.NotEmpty(t, actual.CreatedAt)
					require.Condition(t, func() bool {
						return expected.ID == actual.ID
					})
				},
			)
		})
	}
}

func TestIntegrationNote_Create(t *testing.T) {
	t.Parallel()

	name := gofakeit.Name()
	text := gofakeit.Sentence(5)

	cases := []testutil.IntegrationCase[dto.NoteRequest, dto.NoteResponse]{
		{
			Name: "successful_created",
			TokenFactory: func() string {
				return generateAccessToken(t)
			},
			Req: &dto.NoteRequest{
				Name: name,
				Text: text,
			},
			StatusCode: http.StatusCreated,
			Expected: &dto.NoteResponse{
				Name: name,
				Text: text,
			},
		},
		{
			Name: "bad_request",
			TokenFactory: func() string {
				return generateAccessToken(t)
			},
			Req: &dto.NoteRequest{
				Name: name,
			},
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			Name:       "unauthorized",
			Token:      gofakeit.LetterN(20),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(t, http.MethodPost, "/api/v1/notes", func(expected, actual *dto.NoteResponse) {
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

	cases := []testutil.IntegrationCase[dto.NoteRequest, dto.NoteResponse]{
		{
			Name:       "successful_updated",
			Additional: readyNote.ID,
			Token:      token,
			Req: &dto.NoteRequest{
				Name: newName,
				Text: newText,
			},
			StatusCode: http.StatusOK,
			Expected: &dto.NoteResponse{
				ID:   readyNote.ID,
				Name: newName,
				Text: newText,
			},
		},
		{
			Name:       "bad_id",
			Additional: strconv.Itoa(gofakeit.Int()),
			Token:      token,
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			Name:       "bad_request",
			Additional: readyNote.ID,
			Token:      token,
			Req: &dto.NoteRequest{
				Name: newName,
			},
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
		{
			Name:       "not_found",
			Additional: uuid.Must(uuid.NewV7()),
			Token:      token,
			Req: &dto.NoteRequest{
				Name: gofakeit.Name(),
				Text: gofakeit.Sentence(5),
			},
			StatusCode: http.StatusNotFound,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
		},
		{
			Name:       "unauthorized",
			Additional: uuid.Must(uuid.NewV7()),
			Token:      gofakeit.LetterN(20),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(
				t,
				http.MethodPut,
				fmt.Sprintf("/api/v1/notes/%s", tc.Additional),
				func(expected, actual *dto.NoteResponse) {
					require.Equal(t, expected.Name, actual.Name)
					require.Equal(t, expected.Text, actual.Text)
					require.NotEmpty(t, actual.CreatedAt)
					require.Condition(t, func() bool {
						return expected.ID == actual.ID
					})
				},
			)
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

	cases := []testutil.IntegrationCase[any, dto.NoteResponse]{
		{
			Name:       "successful_deleted",
			Additional: readyNote.ID,
			Token:      token,
			StatusCode: http.StatusOK,
			Expected: &dto.NoteResponse{
				ID:   readyNote.ID,
				Name: readyNote.Name,
				Text: readyNote.Text,
			},
		},
		{
			Name:       "bad_id",
			Additional: strconv.Itoa(gofakeit.Int()),
			Token:      token,
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			Name:       "not_found",
			Additional: uuid.Must(uuid.NewV7()),
			Token:      token,
			StatusCode: http.StatusNotFound,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeNotFound,
				},
			},
		},
		{
			Name:       "unauthorized",
			Additional: uuid.Must(uuid.NewV7()),
			Token:      gofakeit.LetterN(20),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(
				t,
				http.MethodDelete,
				fmt.Sprintf("/api/v1/notes/%s", tc.Additional),
				func(expected, actual *dto.NoteResponse) {
					require.Equal(t, expected.Name, actual.Name)
					require.Equal(t, expected.Text, actual.Text)
					require.NotEmpty(t, actual.CreatedAt)
					require.Condition(t, func() bool {
						return expected.ID == actual.ID
					})

					require.False(t, noteExists(t, token, readyNote.ID))
				},
			)
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

	cases := []testutil.IntegrationCase[search.Request, dto.NoteSearchResponse]{
		{
			Name:  "successful_search_all",
			Token: token,
			Req: &search.Request{
				Orders: []search.Order{
					{Key: "id", Desc: false},
				},
				Limit: uint64(len(notes)),
			},
			StatusCode: http.StatusOK,
			Expected: &dto.NoteSearchResponse{
				TotalRows: uint64(len(notes)),
				Rows:      notes,
			},
		},
		{
			Name:  "successful_search_limited",
			Token: token,
			Req: &search.Request{
				Orders: []search.Order{
					{Key: "id", Desc: false},
				},
				Limit: uint64(len(limitedNotes)),
			},
			StatusCode: http.StatusOK,
			Expected: &dto.NoteSearchResponse{
				TotalRows: uint64(len(notes)),
				Rows:      limitedNotes,
			},
		},
		{
			Name:  "successful_search_filtered",
			Token: token,
			Req: &search.Request{
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
			StatusCode: http.StatusOK,
			Expected: &dto.NoteSearchResponse{
				TotalRows: 2,
				Rows:      notes[:2],
			},
		},
		{
			Name:  "successful_search_filtered_offset",
			Token: token,
			Req: &search.Request{
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
			StatusCode: http.StatusOK,
			Expected: &dto.NoteSearchResponse{
				TotalRows: 2,
				Rows:      notes[1:2],
			},
		},
		{
			Name:  "unavailable_field",
			Token: token,
			Req: &search.Request{
				Filters: map[string]any{
					"unavailable": "1",
				},
			},
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeBadRequest,
				},
			},
		},
		{
			Name:       "bad_request",
			Token:      token,
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeJsonParse,
				},
			},
		},
		{
			Name:       "unauthorized",
			Token:      gofakeit.LetterN(20),
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(t, http.MethodPost, "/api/v1/notes/search", func(expected, actual *dto.NoteSearchResponse) {
				require.Equal(t, tc.Expected.TotalRows, actual.TotalRows)
				require.Len(t, actual.Rows, len(tc.Expected.Rows))

				for i, expected := range tc.Expected.Rows {
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

	httpReq, err := http.NewRequest(http.MethodPost, testutil.WithBaseUrl("/api/v1/notes"), bytes.NewBuffer(jsonReq))
	require.NoError(t, err)
	httpReq.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	defer res.Body.Close() // nolint: errcheck

	require.Equal(t, http.StatusCreated, res.StatusCode)

	n := &dto.NoteResponse{}
	require.NoError(t, json.NewDecoder(res.Body).Decode(n))
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

	httpReq, err := http.NewRequest(http.MethodGet, testutil.WithBaseUrl(fmt.Sprintf("/api/v1/notes/%s", id)), nil)
	require.NoError(t, err)
	httpReq.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(httpReq) // nolint: bodyclose
	require.NoError(t, err)
	defer res.Body.Close() // nolint: errcheck

	require.Condition(t, func() bool {
		return res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusOK
	})

	return http.StatusNotFound != res.StatusCode
}
