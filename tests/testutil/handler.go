package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/middleware"
	"github.com/xsqrty/notes/mocks/app/mock_app"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
)

type HandlerCase[REQ, RES, DEPS any] struct {
	Name        string
	ID          string
	Req         REQ
	StatusCode  int
	Expected    RES
	ExpectedErr *httpio.ErrorResponse
	Mocker      func(REQ, DEPS)
}

func (tc *HandlerCase[REQ, RES, DEPS]) Run(
	t *testing.T,
	method, url string,
	builder func() DEPS,
	handler func(DEPS) http.HandlerFunc,
) {
	t.Helper()
	deps := builder()

	if tc.Mocker != nil {
		tc.Mocker(tc.Req, deps)
	}

	var body io.Reader
	if reflect.ValueOf(tc.Req).Kind() == reflect.Ptr && !reflect.ValueOf(tc.Req).IsNil() {
		b, err := json.Marshal(tc.Req)
		require.NoError(t, err)
		body = bytes.NewReader(b)
	}

	r := httptest.NewRequest(method, url, body)
	w := httptest.NewRecorder()

	middleware.Logger(mock_app.NewDeps(t, func(deps *app.Deps) {}).Logger)(
		handler(deps),
	).ServeHTTP(w, AddUrlParams(r, map[string]string{
		"id": tc.ID,
	}))
	res := w.Result()

	require.Equal(t, tc.StatusCode, res.StatusCode)
	if reflect.ValueOf(tc.Expected).Kind() == reflect.Ptr && !reflect.ValueOf(tc.Expected).IsNil() {
		var result RES
		require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
		require.Equal(t, tc.Expected, result)
	} else {
		err := httpio.ErrorResponse{}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&err))
		require.Equal(t, tc.ExpectedErr.Error.Code, err.Error.Code)
		require.NotEmpty(t, err.Error.Message)
	}
}

func AddUrlParams(r *http.Request, params map[string]string) *http.Request {
	ctx := chi.NewRouteContext()
	for k, v := range params {
		ctx.URLParams.Add(k, v)
	}

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}
