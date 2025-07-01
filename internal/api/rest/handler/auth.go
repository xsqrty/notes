package handler

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/xsqrty/notes/internal/adapter/dtoadapter"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/internal/middleware"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"net/http"
)

type AuthHandler struct {
	deps *app.Deps
}

func NewAuthHandler(deps *app.Deps) *AuthHandler {
	return &AuthHandler{deps}
}

func (h *AuthHandler) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/signup", h.SignUp)
	router.Post("/login", h.Login)
	router.With(h.deps.JWTAuthentication.VerifyRefresh).Post("/refresh", h.RefreshToken)
	return router
}

// Login handler
//
//	@Summary		Login
//	@Description	Login user with email&password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Login request"
//	@Success		201		{object}	dto.TokenResponse
//	@Failure		400		{object}	httpio.ErrorResponse
//	@Failure		401		{object}	httpio.ErrorResponse
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	request, err := httpio.Parse[dto.LoginRequest](http.MaxBytesReader(w, r.Body, int64(h.deps.Config.Server.LimitReqJson)))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("get tokens parse request")
		httpio.Error(w, http.StatusBadRequest, err)
		return
	}

	tokens, err := h.deps.Service.AuthService.Login(r.Context(), dtoadapter.LoginRequestDtoToEntity(&request))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("get tokens")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	httpio.Json(w, http.StatusCreated, dtoadapter.TokensToResponseDto(tokens))
}

// SignUp handler
//
//	@Summary		Sign up
//	@Description	Register a new user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.SignUpRequest	true	"Sign up request"
//	@Success		201		{object}	dto.TokenResponse
//	@Failure		400		{object}	httpio.ErrorResponse
//	@Failure		401		{object}	httpio.ErrorResponse
//	@Failure		500		{object}	httpio.ErrorResponse
//	@Router			/auth/signup [post]
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	request, err := httpio.Parse[dto.SignUpRequest](http.MaxBytesReader(w, r.Body, int64(h.deps.Config.Server.LimitReqJson)))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("signup handler parse request")
		httpio.Error(w, http.StatusBadRequest, err)
		return
	}

	tokens, err := h.deps.Service.AuthService.SignUp(r.Context(), dtoadapter.SignUpRequestDtoToEntity(&request))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("signup handler")
		if errors.Is(err, auth.ErrEmailAlreadyExists) {
			httpio.Error(w, http.StatusBadRequest, errx.New(errx.CodeEmailExists, "Email already exists"))
		} else {
			httpio.Error(w, http.StatusInternalServerError, err)
		}
		return
	}

	httpio.Json(w, http.StatusCreated, dtoadapter.TokensToResponseDto(tokens))
}

// RefreshToken handler
//
//	@Summary		Refresh token
//	@Description	Create new tokens based on refresh token
//	@Tags			Auth
//	@Produce		json
//	@Param			Authorization	header		string	true	"Bearer {YOUR REFRESH TOKEN}"
//	@Success		201				{object}	dto.TokenResponse
//	@Failure		400				{object}	httpio.ErrorResponse
//	@Failure		401				{object}	httpio.ErrorResponse
//	@Router			/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	user, err := h.deps.JWTAuthentication.GetUser(r)
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("refresh token get user")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	tokens, err := h.deps.Service.AuthService.GenerateTokens(user)
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("refresh get token")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	httpio.Json(w, http.StatusCreated, dtoadapter.TokensToResponseDto(tokens))
}
