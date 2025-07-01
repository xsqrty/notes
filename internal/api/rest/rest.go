package rest

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/xsqrty/notes/internal/api/rest/handler"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/middleware"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"net/http"
)

const Entrypoint = "/api/v1"

type Rest interface {
	Routes() *chi.Mux
}

type rest struct {
	deps *app.Deps
}

// NewRest create application rest service
//
//	@title						Note API
//	@version					1.0
//	@description				This is a sample note service
//	@contact.name				XSQRTY
//	@contact.url				https://github.com/xsqrty/
//	@contact.email				tbq.active@gmail.com
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	AccessTokenAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer {YOUR TOKEN}" to correctly set the API Key
//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
func NewRest(deps *app.Deps) Rest {
	return &rest{deps}
}

func (r *rest) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Mount("/auth", handler.NewAuthHandler(r.deps).Routes())
	router.With(r.deps.JWTAuthentication.Verify).Mount("/notes", handler.NewNoteHandler(r.deps).Routes())

	entrypoint := chi.NewRouter()
	entrypoint.Use(cors.Handler(cors.Options{
		AllowedOrigins:   r.deps.Config.Cors.AllowedOrigins,
		AllowedMethods:   r.deps.Config.Cors.AllowedMethods,
		AllowedHeaders:   r.deps.Config.Cors.AllowedHeaders,
		AllowCredentials: r.deps.Config.Cors.AllowCredentials,
		MaxAge:           r.deps.Config.Cors.MaxAge,
	}))

	entrypoint.Use(middleware.Metrics(r.deps.Metrics.Http))
	entrypoint.Use(middleware.RequestID)
	entrypoint.Use(middleware.Logger(r.deps.Logger))
	entrypoint.Use(middleware.Recover)
	entrypoint.Mount(Entrypoint, router)

	entrypoint.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpio.Error(w, http.StatusMethodNotAllowed, errx.New(errx.CodeMethodNotAllowed, "Method not allowed"))
	})

	entrypoint.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpio.Error(w, http.StatusNotFound, errx.New(errx.CodeNotFound, fmt.Sprintf("%s %s not found", r.Method, r.URL.RequestURI())))
	})

	return entrypoint
}
