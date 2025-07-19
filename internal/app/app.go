package app

import (
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/role"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/internal/guards"
	"github.com/xsqrty/notes/internal/logger"
	"github.com/xsqrty/notes/internal/metrics"
	"github.com/xsqrty/notes/internal/middleware"
	"github.com/xsqrty/notes/internal/repository"
	"github.com/xsqrty/notes/internal/service"
	"github.com/xsqrty/notes/pkg/passwd"
	"github.com/xsqrty/op/db"
)

// Deps is a container for application-wide dependencies required by various components.
type Deps struct {
	Logger            *logger.Logger
	Config            *config.Config
	JWTAuthentication middleware.JWTAuthentication
	Repository        ReposSet
	Service           ServicesSet
	Metrics           appMetrics
}

// appMetrics is a structure that holds metrics-related data for the application.
type appMetrics struct {
	Http *metrics.HttpMetrics
}

// ReposSet contains the main repositories used by the application.
type ReposSet struct {
	RoleRepository role.Repository
	UserRepository user.Repository
	NoteRepository note.Repository
}

// ServicesSet contains the main services used by the application.
type ServicesSet struct {
	AuthService auth.Service
	NoteService note.Service
}

// NewDeps initializes and returns a Deps struct populated with configuration, logger, repositories, services, and metrics.
func NewDeps(config *config.Config, log *logger.Logger, pool db.ConnPool) *Deps {
	roleRepo := repository.NewRoleRepository(pool)
	userRepo := repository.NewUserRepo(pool)
	noteRepo := repository.NewNoteRepo(pool)

	jwtAuth := middleware.NewJWTAuthentication(&config.Auth, userRepo)
	passGenerator := passwd.NewPasswordGenerator(config.Auth.PasswordCost)

	return &Deps{
		Logger:            log,
		Config:            config,
		JWTAuthentication: jwtAuth,
		Repository: ReposSet{
			RoleRepository: roleRepo,
			UserRepository: userRepo,
			NoteRepository: noteRepo,
		},
		Service: ServicesSet{
			AuthService: service.NewAuthService(&service.AuthServiceDeps{
				TxManager: pool,
				RoleRepo:  roleRepo,
				UserRepo:  userRepo,
				Tokenizer: jwtAuth,
				PassGen:   passGenerator,
			}),
			NoteService: service.NewNoteService(&service.NoteServiceDeps{
				NoteRepo:  noteRepo,
				NoteGuard: guards.NewNoteGuarder(roleRepo),
			}),
		},
		Metrics: appMetrics{
			Http: metrics.NewHttpMetrics(config.Metrics),
		},
	}
}

// Close releases Deps resources.
func (d *Deps) Close() error {
	return d.JWTAuthentication.Close()
}
