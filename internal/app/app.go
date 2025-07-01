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

type Deps struct {
	Logger            *logger.Logger
	Config            *config.Config
	JWTAuthentication middleware.JWTAuthentication
	Repository        ReposSet
	Service           ServicesSet
	Metrics           appMetrics
}

type appMetrics struct {
	Http *metrics.HttpMetrics
}

type ReposSet struct {
	RoleRepository role.Repository
	UserRepository user.Repository
	NoteRepository note.Repository
}

type ServicesSet struct {
	AuthService auth.Service
	NoteService note.Service
}

func NewDeps(config *config.Config, log *logger.Logger, pool db.ConnPool) *Deps {
	roleRepo := repository.NewRoleRepository(pool)
	userRepo := repository.NewUserRepo(pool)
	noteRepo := repository.NewNoteRepo(pool)

	jwtAuth := middleware.NewJWTAuthentication("user_id", &config.Auth, userRepo)
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

func (d *Deps) Close() error {
	return d.JWTAuthentication.Close()
}
