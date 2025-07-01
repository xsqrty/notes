package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"net/http"
	"time"
)

type HealthCheckHandler struct {
	deps *app.Deps
}

func NewHealthCheckHandler(deps *app.Deps) *HealthCheckHandler {
	return &HealthCheckHandler{deps}
}

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
//	@Success		200		{object}	dto.HealthCheckResponse
//	@Router			/healthcheck [get]
func (h *HealthCheckHandler) Healthcheck(w http.ResponseWriter, _ *http.Request) {
	httpio.Json(w, http.StatusOK, &dto.HealthCheckResponse{
		Version:     h.deps.Config.Version,
		AppName:     h.deps.Config.AppName,
		CurrentTime: time.Now().UnixMilli(),
	})
}
