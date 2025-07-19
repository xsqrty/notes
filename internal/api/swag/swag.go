package swag

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/xsqrty/notes/docs"
	"github.com/xsqrty/notes/internal/config"
)

// NewSwagServer initializes and returns a new HTTP server configured for serving Swagger documentation.
// It sets up routing to handle requests to the root and Swagger UI.
// The server's address is determined by the given SwagConfig.
func NewSwagServer(cfg config.SwagConfig) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
	})

	router.Get("/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 10,
	}
}
