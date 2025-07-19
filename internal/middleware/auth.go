package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/pkg/jwtsafe"
)

// JWTAuthentication defines methods for managing JWT authentication, including token generation and verification.
type JWTAuthentication interface {
	Close() error
	CreateAccessToken(user *user.User) (string, error)
	CreateRefreshToken(user *user.User) (string, error)
	GetUser(r *http.Request) (*user.User, error)
	Verify(next http.Handler) http.Handler
	VerifyRefresh(next http.Handler) http.Handler
}

// jwtAuthentication is an internal implementation of the JWTAuthentication interface using JWTSafe for token management.
// It facilitates the encoding, decoding, and verification of JWTs for handling access and refresh token workflows.
type jwtAuthentication struct {
	repo       user.Repository
	accessJwt  jwtsafe.JWTSafe
	refreshJwt jwtsafe.JWTSafe
}

// userKeyType represents a custom type based on string, typically used for defining keys related to user-specific data.
type userKeyType string

var (
	ErrUnauthorized   = errx.New(errx.CodeUnauthorized, "user not authorized")
	ErrTokenExpired   = errx.New(errx.CodeTokenExpired, "token is expired")
	ErrBearerRequired = errx.New(errx.CodeUnauthorized, "bearer token is required")
)

const (
	// secretSize defines the size of the secret key used for encryption or signing.
	secretSize = 32
	// userIDClaimKey represents the claim key for storing the user ID in a token.
	userIDClaimKey = "user_id"
	// userKey is a key type used for identifying user-related context values.
	userKey = userKeyType("user_id")
)

// NewJWTAuthentication initializes and returns a JWTAuthentication implementation.
func NewJWTAuthentication(authConf *config.AuthConfig, repo user.Repository) JWTAuthentication {
	return &jwtAuthentication{
		repo:       repo,
		accessJwt:  jwtsafe.New(authConf.AccessTokenExp, secretSize),
		refreshJwt: jwtsafe.New(authConf.RefreshTokenExp, secretSize),
	}
}

// Close releases resources held by accessJwt and refreshJwt; returns an error if any operation fails.
func (j *jwtAuthentication) Close() error {
	if err := j.accessJwt.Close(); err != nil {
		return err
	}

	if err := j.refreshJwt.Close(); err != nil {
		return err
	}

	return nil
}

// CreateAccessToken generates a new access token for the provided user and returns it as a string.
// Returns an error if token encoding fails.
func (j *jwtAuthentication) CreateAccessToken(user *user.User) (string, error) {
	return j.accessJwt.Encode(jwtsafe.MapClaims{userIDClaimKey: user.ID})
}

// CreateRefreshToken generates a new refresh token for the given user and returns it as a string.
func (j *jwtAuthentication) CreateRefreshToken(user *user.User) (string, error) {
	return j.refreshJwt.Encode(jwtsafe.MapClaims{userIDClaimKey: user.ID})
}

// GetUser retrieves a user from the request context using the extracted user ID and fetches the user details from the repository.
func (j *jwtAuthentication) GetUser(r *http.Request) (*user.User, error) {
	idCtx := r.Context().Value(userKey)
	if idCtx == nil {
		return nil, fmt.Errorf("ctx doesn't have user_id: %w", ErrUnauthorized)
	}

	id, err := uuid.Parse(idCtx.(string))
	if err != nil {
		return nil, fmt.Errorf("incorrect user_id: %w", err)
	}

	user, err := j.repo.GetByID(r.Context(), id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// Verify is a middleware that validates the access JWT and adds the user ID to the request context if valid.
func (j *jwtAuthentication) Verify(next http.Handler) http.Handler {
	return j.verify(j.accessJwt)(next)
}

// VerifyRefresh authenticates HTTP requests using a refresh JWT for token validation.
func (j *jwtAuthentication) VerifyRefresh(next http.Handler) http.Handler {
	return j.verify(j.refreshJwt)(next)
}

// verify creates middleware to validate JWT tokens and inject the user ID from the claims into the request context.
func (j *jwtAuthentication) verify(jwt jwtsafe.JWTSafe) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				httpio.Error(w, http.StatusUnauthorized, ErrBearerRequired)
				return
			}

			claims, err := jwt.Decode(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				if errors.Is(err, jwtsafe.ErrJWTExpired) {
					httpio.Error(w, http.StatusUnauthorized, ErrTokenExpired)
				} else {
					httpio.Error(w, http.StatusUnauthorized, ErrUnauthorized)
				}

				return
			}

			if userId, ok := claims[userIDClaimKey].(string); ok {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, userId)))
				return
			}

			httpio.Error(w, http.StatusUnauthorized, ErrUnauthorized)
		})
	}
}
