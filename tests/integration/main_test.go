package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/xsqrty/notes/internal/api/rest"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/internal/logger"
	"github.com/xsqrty/notes/mocks/app/mock_app"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/op/db"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	postgresUser = "postgres"
	postgresPass = "postgres"
	postgresDB   = "integration_test"
)

type integrationCase[REQ, RES any] struct {
	name         string
	req          *REQ
	token        string
	tokenFactory func() string
	statusCode   int
	expectedErr  *httpio.ErrorResponse
	expected     *RES
	additional   any
	onSuccess    func()
}

var ctx = context.Background()
var server *httptest.Server
var rootTokens *dto.TokenResponse

var (
	rootEmail    = gofakeit.Email()
	rootPassword = gofakeit.Password(true, true, true, true, true, 20)
	rootName     = gofakeit.Name()
)

func TestMain(m *testing.M) {
	dsn, cleanup, err := startPostgresContainer(ctx)
	defer cleanup()
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	pool, err := db.OpenPostgres(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	err = mock_app.AutoMigrate("file://../../migrations", dsn)
	if err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	deps := app.NewDeps(cfg, &logger.Logger{
		Logger: zerolog.Nop(),
	}, pool)
	defer deps.Close()

	server = httptest.NewServer(rest.NewRest(deps).Routes())
	defer server.Close()

	rootTokens, err = signUpUser(&dto.SignUpRequest{
		Email:    rootEmail,
		Password: rootPassword,
		Name:     rootName,
	})

	if err != nil {
		log.Fatalf("failed to sign up user: %v", err)
	}

	m.Run()
}

func withBaseUrl(url string) string {
	return fmt.Sprintf("%s%s", server.URL, url)
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
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgresUser, postgresPass, host, port.Port(), postgresDB)
	return dsn, func() {
		container.Terminate(ctx)
	}, nil
}

func signUpUser(req *dto.SignUpRequest) (*dto.TokenResponse, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(withBaseUrl("/api/v1/auth/signup"), "application/json", bytes.NewReader(jsonReq))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

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

func (tc *integrationCase[REQ, RES]) run(t *testing.T, method, url string, checker func(expected, actual *RES)) {
	t.Helper()
	var jsonReq []byte

	if tc.req != nil {
		var err error
		jsonReq, err = json.Marshal(tc.req)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, withBaseUrl(url), bytes.NewReader(jsonReq))
	require.NoError(t, err)

	token := tc.token
	if tc.tokenFactory != nil {
		token = tc.tokenFactory()
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, tc.statusCode, res.StatusCode)

	if tc.expected != nil {
		n := new(RES)
		json.NewDecoder(res.Body).Decode(n)
		checker(tc.expected, n)
		if tc.onSuccess != nil {
			tc.onSuccess()
		}
	} else {
		err := &httpio.ErrorResponse{}
		json.NewDecoder(res.Body).Decode(err)
		require.Equal(t, tc.expectedErr.Error.Code, err.Error.Code)
		require.NotEmpty(t, err.Error.Message)
	}
}
