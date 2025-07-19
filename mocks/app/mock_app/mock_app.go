package mock_app

import (
	"net/url"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/internal/logger"
	"github.com/xsqrty/notes/mocks/domain/mock_auth"
	"github.com/xsqrty/notes/mocks/domain/mock_note"
	"github.com/xsqrty/notes/pkg/config/size"
)

func NewDeps(t *testing.T, mocker func(deps *app.Deps)) *app.Deps {
	deps := &app.Deps{
		Logger: &logger.Logger{
			Logger: zerolog.Nop(),
		},
		Config: &config.Config{
			Server: config.ServerConfig{
				LimitReqJson: size.Bytes((1 << 10) * 100), // 100 kb
			},
		},
		Service: app.ServicesSet{
			AuthService: mock_auth.NewService(t),
			NoteService: mock_note.NewService(t),
		},
	}

	mocker(deps)
	return deps
}

func AutoMigrate(source, dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return err
	}

	q := parsed.Query()
	q.Set("sslmode", "disable")
	parsed.RawQuery = q.Encode()

	m, err := migrate.New(source, parsed.String())
	if err != nil {
		return err
	}

	return m.Up()
}
