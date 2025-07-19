package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/tests/testutil"
)

func TestIntegrationAuth_Login(t *testing.T) {
	t.Parallel()
	cases := []testutil.IntegrationCase[dto.LoginRequest, dto.TokenResponse]{
		{
			Name: "successful_login",
			Req: &dto.LoginRequest{
				Email:    rootEmail,
				Password: rootPassword,
			},
			StatusCode: http.StatusCreated,
			Expected: &dto.TokenResponse{
				User: &dto.UserResponse{
					Name:  rootName,
					Email: rootEmail,
				},
			},
		},
		{
			Name: "user_not_found",
			Req: &dto.LoginRequest{
				Email:    gofakeit.Email(),
				Password: rootPassword,
			},
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
		{
			Name: "validation_error",
			Req: &dto.LoginRequest{
				Email:    gofakeit.Name(),
				Password: rootPassword,
			},
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(t, http.MethodPost, "/api/v1/auth/login", func(expected, actual *dto.TokenResponse) {
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
				require.Equal(t, expected.User.Name, actual.User.Name)
				require.Equal(t, expected.User.Email, actual.User.Email)
			})
		})
	}
}

func TestIntegrationAuth_SignUp(t *testing.T) {
	t.Parallel()

	email := gofakeit.Email()
	name := gofakeit.Name()
	password := gofakeit.Password(true, true, true, true, true, 10)

	cases := []testutil.IntegrationCase[dto.SignUpRequest, dto.TokenResponse]{
		{
			Name: "successful_signup",
			Req: &dto.SignUpRequest{
				Name:     name,
				Email:    email,
				Password: password,
			},
			StatusCode: http.StatusCreated,
			Expected: &dto.TokenResponse{
				User: &dto.UserResponse{
					Name:  name,
					Email: email,
				},
			},
			OnSuccess: func() {
				require.NotEmpty(t, login(t, &dto.LoginRequest{
					Email:    email,
					Password: password,
				}).AccessToken)
			},
		},
		{
			Name: "email_already_exists",
			Req: &dto.SignUpRequest{
				Name:     gofakeit.Name(),
				Email:    rootEmail,
				Password: rootPassword,
			},
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeEmailExists,
				},
			},
		},
		{
			Name: "validation_error",
			Req: &dto.SignUpRequest{
				Email:    gofakeit.Name(),
				Password: rootPassword,
			},
			StatusCode: http.StatusBadRequest,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeValidation,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(t, http.MethodPost, "/api/v1/auth/signup", func(expected, actual *dto.TokenResponse) {
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
				require.Equal(t, expected.User.Name, actual.User.Name)
				require.Equal(t, expected.User.Email, actual.User.Email)
			})
		})
	}
}

func TestIntegrationAuth_RefreshToken(t *testing.T) {
	t.Parallel()

	cases := []testutil.IntegrationCase[any, dto.TokenResponse]{
		{
			Name:       "successful_refresh",
			Token:      rootTokens.RefreshToken,
			StatusCode: http.StatusCreated,
			Expected: &dto.TokenResponse{
				User: &dto.UserResponse{
					Name:  rootName,
					Email: rootEmail,
				},
			},
		},
		{
			Name:       "incorrect_token",
			Token:      rootTokens.AccessToken,
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
		{
			Name:       "bad_request",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			ExpectedErr: &httpio.ErrorResponse{
				Error: &errx.CodeError{
					Code: errx.CodeUnauthorized,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(t, http.MethodPost, "/api/v1/auth/refresh", func(expected, actual *dto.TokenResponse) {
				require.NotEmpty(t, actual.AccessToken)
				require.NotEmpty(t, actual.RefreshToken)
				require.Equal(t, expected.User.Name, actual.User.Name)
				require.Equal(t, expected.User.Email, actual.User.Email)
			})
		})
	}
}

func login(t *testing.T, req *dto.LoginRequest) *dto.TokenResponse {
	t.Helper()
	jsonReq, err := json.Marshal(req)
	require.NoError(t, err)

	res, err := http.Post(testutil.WithBaseUrl("/api/v1/auth/login"), "application/json", bytes.NewReader(jsonReq))
	require.NoError(t, err)
	defer res.Body.Close() // nolint: errcheck

	require.Equal(t, http.StatusCreated, res.StatusCode)

	tokens := &dto.TokenResponse{}
	require.NoError(t, json.NewDecoder(res.Body).Decode(tokens))
	require.Equal(t, req.Email, tokens.User.Email)

	return tokens
}

func signUp(t *testing.T, req *dto.SignUpRequest) *dto.TokenResponse {
	t.Helper()
	jsonReq, err := json.Marshal(req)
	require.NoError(t, err)

	res, err := http.Post(testutil.WithBaseUrl("/api/v1/auth/signup"), "application/json", bytes.NewReader(jsonReq))
	require.NoError(t, err)
	defer res.Body.Close() // nolint: errcheck

	require.Equal(t, http.StatusCreated, res.StatusCode)

	tokens := &dto.TokenResponse{}
	require.NoError(t, json.NewDecoder(res.Body).Decode(tokens))
	require.Equal(t, req.Email, tokens.User.Email)

	return tokens
}
