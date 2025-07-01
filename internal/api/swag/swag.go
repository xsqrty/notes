package swag

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/swaggo/http-swagger"
	_ "github.com/xsqrty/notes/docs"
	"github.com/xsqrty/notes/internal/config"
	"net/http"
)

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
		Addr:    addr,
		Handler: router,
	}
}
