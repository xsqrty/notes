package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/xsqrty/notes/internal/api/rest"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/internal/logger"
	"github.com/xsqrty/notes/mocks/app/mock_app"
	"github.com/xsqrty/notes/tests/testutil"
	"github.com/xsqrty/op/db/postgres"
)

const (
	postgresUser = "postgres"
	postgresPass = "postgres"
	postgresDB   = "integration_test"
)

var (
	ctx        = context.Background()
	rootTokens *dto.TokenResponse
)

var (
	rootEmail    = gofakeit.Email()
	rootPassword = gofakeit.Password(true, true, true, true, true, 20)
	rootName     = gofakeit.Name()
)

func TestMain(m *testing.M) {
	dsn, cleanup, err := startPostgresContainer(ctx)
	if err != nil {
		log.Panicf("failed to start container: %v", err)
	}
	defer cleanup()

	pool, err := postgres.Open(ctx, dsn)
	if err != nil {
		log.Panicf("failed to connect to postgres: %v", err)
	}

	err = mock_app.AutoMigrate("file://../../migrations", dsn)
	if err != nil {
		log.Panicf("failed to auto migrate: %v", err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	deps := app.NewDeps(cfg, &logger.Logger{
		Logger: zerolog.Nop(),
	}, pool)
	defer deps.Close() // nolint: errcheck

	testutil.Server = httptest.NewServer(rest.NewRest(deps).Routes())
	defer testutil.Server.Close()

	rootTokens, err = signUpUser(&dto.SignUpRequest{
		Email:    rootEmail,
		Password: rootPassword,
		Name:     rootName,
	})
	if err != nil {
		log.Panicf("failed to sign up user: %v", err)
	}

	m.Run()
}

func startPostgresContainer(ctx context.Context) (string, func(), error) {
	req := testcontainers.ContainerRequest{
		Name:         "postgres",
		Image:        "postgres:17.5-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     postgresUser,
			"POSTGRES_PASSWORD": postgresPass,
			"POSTGRES_DB":       postgresDB,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return "", nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return "", nil, err
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresUser,
		postgresPass,
		host,
		port.Port(),
		postgresDB,
	)
	return dsn, func() {
		container.Terminate(ctx) // nolint: gosec, errcheck
	}, nil
}

func signUpUser(req *dto.SignUpRequest) (*dto.TokenResponse, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(testutil.WithBaseUrl("/api/v1/auth/signup"), "application/json", bytes.NewReader(jsonReq))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("invalid response code: %d", res.StatusCode)
	}

	tokens := &dto.TokenResponse{}
	return tokens, json.NewDecoder(res.Body).Decode(tokens)
}

func generateAccessToken(t *testing.T) string {
	return signUp(t, &dto.SignUpRequest{
		Name:     gofakeit.Name(),
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 20),
	}).AccessToken
}
