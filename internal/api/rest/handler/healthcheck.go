package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
)

// HealthCheckHandler provides an HTTP handler for managing application health checks.
type HealthCheckHandler struct {
	deps *app.Deps
}

// NewHealthCheckHandler initializes and returns a new instance of HealthCheckHandler with provided dependencies.
func NewHealthCheckHandler(deps *app.Deps) *HealthCheckHandler {
	return &HealthCheckHandler{deps}
}

// Routes configure and return a new HTTP router for handling health check requests.
func (h *HealthCheckHandler) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", h.Healthcheck)
	return router
}

// Healthcheck handler
//
//	@Summary		Healthcheck
//	@Description	Get application version, name, current time
//	@Tags			Healthcheck
//	@Produce		json
//	@Success		200	{object}	dto.HealthCheckResponse
//	@Router			/healthcheck [get]
func (h *HealthCheckHandler) Healthcheck(w http.ResponseWriter, _ *http.Request) {
	httpio.Json(w, http.StatusOK, &dto.HealthCheckResponse{
		Version:     h.deps.Config.Version,
		AppName:     h.deps.Config.AppName,
		CurrentTime: time.Now().UnixMilli(),
	})
}
