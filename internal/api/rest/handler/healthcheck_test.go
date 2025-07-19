package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/mocks/app/mock_app"
)

func TestHealthCheckHandler_Healthcheck(t *testing.T) {
	t.Parallel()

	version := gofakeit.AppVersion()
	name := gofakeit.AppName()

	handler := NewHealthCheckHandler(mock_app.NewDeps(t, func(deps *app.Deps) {
		deps.Config.AppName = name
		deps.Config.Version = version
	}))

	r := httptest.NewRequest(http.MethodPost, "/api/v1/healthcheck", nil)
	w := httptest.NewRecorder()

	handler.Healthcheck(w, r)
	res := w.Result()

	require.Equal(t, http.StatusOK, res.StatusCode)

	result := dto.HealthCheckResponse{}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	require.Equal(t, version, result.Version)
	require.Equal(t, name, result.AppName)
	require.GreaterOrEqual(t, time.Now().UnixMilli(), result.CurrentTime)
}
