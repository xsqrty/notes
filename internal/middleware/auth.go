package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/config"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"github.com/xsqrty/notes/pkg/jwtsafe"
	"net/http"
	"strings"
)

var (
	ErrUnauthorized   = errx.New(errx.CodeUnauthorized, "user not authorized")
	ErrTokenExpired   = errx.New(errx.CodeTokenExpired, "token is expired")
	ErrBearerRequired = errx.New(errx.CodeUnauthorized, "bearer token is required")
)

const (
	secretSize = 32
)

type JWTAuthentication interface {
	Close() error
	CreateAccessToken(user *user.User) (string, error)
	CreateRefreshToken(user *user.User) (string, error)
	GetUser(r *http.Request) (*user.User, error)
	Verify(next http.Handler) http.Handler
	VerifyRefresh(next http.Handler) http.Handler
}

type jwtAuthentication struct {
	repo       user.Repository
	userIdKey  string
	accessJwt  jwtsafe.JWTSafe
	refreshJwt jwtsafe.JWTSafe
}

func NewJWTAuthentication(userKey string, authConf *config.AuthConfig, repo user.Repository) JWTAuthentication {
	return &jwtAuthentication{
		repo:       repo,
		userIdKey:  userKey,
		accessJwt:  jwtsafe.New(authConf.AccessTokenExp, secretSize),
		refreshJwt: jwtsafe.New(authConf.RefreshTokenExp, secretSize),
	}
}

func (j *jwtAuthentication) Close() error {
	if err := j.accessJwt.Close(); err != nil {
		return err
	}

	if err := j.refreshJwt.Close(); err != nil {
		return err
	}

	return nil
}

func (j *jwtAuthentication) CreateAccessToken(user *user.User) (string, error) {
	return j.accessJwt.Encode(jwtsafe.MapClaims{j.userIdKey: user.ID})
}

func (j *jwtAuthentication) CreateRefreshToken(user *user.User) (string, error) {
	return j.refreshJwt.Encode(jwtsafe.MapClaims{j.userIdKey: user.ID})
}

func (j *jwtAuthentication) GetUser(r *http.Request) (*user.User, error) {
	idCtx := r.Context().Value(j.userIdKey)
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

func (j *jwtAuthentication) Verify(next http.Handler) http.Handler {
	return j.verify(j.accessJwt)(next)
}

func (j *jwtAuthentication) VerifyRefresh(next http.Handler) http.Handler {
	return j.verify(j.refreshJwt)(next)
}

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

			if userId, ok := claims[j.userIdKey].(string); ok {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), j.userIdKey, userId)))
				return
			}

			httpio.Error(w, http.StatusUnauthorized, ErrUnauthorized)
		})
	}
}
