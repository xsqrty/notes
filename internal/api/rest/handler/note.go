package handler

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/xsqrty/notes/internal/adapter/dtoadapter"
	"github.com/xsqrty/notes/internal/app"
	"github.com/xsqrty/notes/internal/domain/note"
	"github.com/xsqrty/notes/internal/domain/search"
	"github.com/xsqrty/notes/internal/dto"
	"github.com/xsqrty/notes/internal/middleware"
	"github.com/xsqrty/notes/pkg/httputil/httpio"
	"github.com/xsqrty/notes/pkg/httputil/httpio/errx"
	"net/http"
)

type NoteHandler struct {
	deps *app.Deps
}

func NewNoteHandler(deps *app.Deps) *NoteHandler {
	return &NoteHandler{deps}
}

func (h *NoteHandler) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/", h.Create)
	router.Post("/search", h.Search)
	router.Get("/{id}", h.Get)
	router.Put("/{id}", h.Update)
	router.Delete("/{id}", h.Delete)
	return router
}

// Get handler
//
//	@Summary		Get note
//	@Description	Get note by id
//	@Tags			Notes
//	@Produce		json
//	@Param			id	path		string	true	"Note id"
//	@Success		200	{object}	dto.NoteResponse
//	@Failure		400	{object}	httpio.ErrorResponse
//	@Failure		401	{object}	httpio.ErrorResponse
//	@Failure		404	{object}	httpio.ErrorResponse
//	@Failure		500	{object}	httpio.ErrorResponse
//	@Security		AccessTokenAuth
//	@Router			/notes/{id} [get]
func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := h.deps.JWTAuthentication.GetUser(r)
	if err != nil {
		middleware.Log(r).Error().Err(err).Msg("get note handler unauthorized")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("get note handler parse id")
		httpio.Error(w, http.StatusBadRequest, errx.New(errx.CodeBadRequest, "Bad request"))
		return
	}

	n, err := h.deps.Service.NoteService.Get(r.Context(), user, id)
	if err != nil {
		if errors.Is(err, note.ErrNoteOperationForbiddenForUser) {
			middleware.Log(r).Error().Err(err).Msg("get note forbidden")
			httpio.Error(w, http.StatusForbidden, errx.New(errx.CodeForbidden, "Operation disallowed"))
			return
		}

		if errors.Is(err, note.ErrNoteNotFound) {
			middleware.Log(r).Debug().Err(err).Msg("get note handler not found")
			httpio.Error(w, http.StatusNotFound, errx.New(errx.CodeNotFound, "Note is not found"))
			return
		}

		middleware.Log(r).Error().Err(err).Msg("couldn't get note")
		httpio.Error(w, http.StatusInternalServerError, err)
		return
	}

	httpio.Json(w, http.StatusOK, dtoadapter.NoteToResponseDto(n))
}

// Create handler
//
//	@Summary		Create note
//	@Description	Create new note
//	@Tags			Notes
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.NoteRequest	true	"Create note request"
//	@Success		201		{object}	dto.NoteResponse
//	@Failure		400		{object}	httpio.ErrorResponse
//	@Failure		401		{object}	httpio.ErrorResponse
//	@Failure		500		{object}	httpio.ErrorResponse
//	@Security		AccessTokenAuth
//	@Router			/notes [post]
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, err := h.deps.JWTAuthentication.GetUser(r)
	if err != nil {
		middleware.Log(r).Error().Err(err).Msg("create note handler unauthorized")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	request, err := httpio.Parse[dto.NoteRequest](http.MaxBytesReader(w, r.Body, int64(h.deps.Config.Server.LimitReqJson)))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("create note handler parse request")
		httpio.Error(w, http.StatusBadRequest, err)
		return
	}

	n, err := h.deps.Service.NoteService.Create(r.Context(), user, dtoadapter.NoteRequestDtoToCreateData(&request))
	if err != nil {
		if errors.Is(err, note.ErrNoteOperationForbiddenForUser) {
			middleware.Log(r).Error().Err(err).Msg("create note forbidden")
			httpio.Error(w, http.StatusForbidden, errx.New(errx.CodeForbidden, "Operation disallowed"))
			return
		}

		middleware.Log(r).Error().Err(err).Msg("couldn't create note")
		httpio.Error(w, http.StatusInternalServerError, err)
		return
	}

	httpio.Json(w, http.StatusCreated, dtoadapter.NoteToResponseDto(n))
}

// Update handler
//
//	@Summary		Create note
//	@Description	Create new note
//	@Tags			Notes
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Note id"
//	@Param			request	body		dto.NoteRequest	true	"Create note request"
//	@Success		200		{object}	dto.NoteResponse
//	@Failure		400		{object}	httpio.ErrorResponse
//	@Failure		401		{object}	httpio.ErrorResponse
//	@Failure		500		{object}	httpio.ErrorResponse
//	@Security		AccessTokenAuth
//	@Router			/notes/{id} [put]
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, err := h.deps.JWTAuthentication.GetUser(r)
	if err != nil {
		middleware.Log(r).Error().Err(err).Msg("update note handler unauthorized")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("update note handler parse id")
		httpio.Error(w, http.StatusBadRequest, errx.New(errx.CodeBadRequest, "Bad request"))
		return
	}

	request, err := httpio.Parse[dto.NoteRequest](http.MaxBytesReader(w, r.Body, int64(h.deps.Config.Server.LimitReqJson)))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("update note handler parse request")
		httpio.Error(w, http.StatusBadRequest, err)
		return
	}

	n, err := h.deps.Service.NoteService.Update(r.Context(), user, dtoadapter.NoteRequestDtoToUpdateData(id, &request))
	if err != nil {
		if errors.Is(err, note.ErrNoteOperationForbiddenForUser) {
			middleware.Log(r).Error().Err(err).Msg("update note forbidden")
			httpio.Error(w, http.StatusForbidden, errx.New(errx.CodeForbidden, "Operation disallowed"))
			return
		}

		if errors.Is(err, note.ErrNoteNotFound) {
			middleware.Log(r).Debug().Err(err).Msg("update note handler not found")
			httpio.Error(w, http.StatusNotFound, errx.New(errx.CodeNotFound, "Note is not found"))
			return
		}

		middleware.Log(r).Error().Err(err).Msg("couldn't update note")
		httpio.Error(w, http.StatusInternalServerError, err)
		return
	}

	httpio.Json(w, http.StatusOK, dtoadapter.NoteToResponseDto(n))
}

// Delete handler
//
//	@Summary		Delete note
//	@Description	Delete note by id
//	@Tags			Notes
//	@Produce		json
//	@Param			id	path		string	true	"Note id"
//	@Success		200	{object}	dto.NoteResponse
//	@Failure		400	{object}	httpio.ErrorResponse
//	@Failure		401	{object}	httpio.ErrorResponse
//	@Failure		404	{object}	httpio.ErrorResponse
//	@Failure		500	{object}	httpio.ErrorResponse
//	@Security		AccessTokenAuth
//	@Router			/notes/{id} [delete]
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, err := h.deps.JWTAuthentication.GetUser(r)
	if err != nil {
		middleware.Log(r).Error().Err(err).Msg("delete note handler unauthorized")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("delete note handler parse id")
		httpio.Error(w, http.StatusBadRequest, errx.New(errx.CodeBadRequest, "Bad request"))
		return
	}

	n, err := h.deps.Service.NoteService.Delete(r.Context(), user, id)
	if err != nil {
		if errors.Is(err, note.ErrNoteOperationForbiddenForUser) {
			middleware.Log(r).Error().Err(err).Msg("delete note forbidden")
			httpio.Error(w, http.StatusForbidden, errx.New(errx.CodeForbidden, "Operation disallowed"))
			return
		}

		if errors.Is(err, note.ErrNoteNotFound) {
			middleware.Log(r).Debug().Err(err).Msg("delete note handler not found")
			httpio.Error(w, http.StatusNotFound, errx.New(errx.CodeNotFound, "Note is not found"))
			return
		}

		middleware.Log(r).Error().Err(err).Msg("couldn't delete note")
		httpio.Error(w, http.StatusInternalServerError, err)
		return
	}

	httpio.Json(w, http.StatusOK, dtoadapter.NoteToResponseDto(n))
}

// Search handler
//
//	@Summary		Search notes
//	@Description	Search notes (filtering, ordering, limit, offset)
//	@Tags			Notes
//	@Accept			json
//	@Produce		json
//	@Param			request	body		search.Request	true	"Search request"
//	@Success		200		{object}	dto.NoteSearchResponse
//	@Failure		400		{object}	httpio.ErrorResponse
//	@Failure		401		{object}	httpio.ErrorResponse
//	@Failure		500		{object}	httpio.ErrorResponse
//	@Security		AccessTokenAuth
//	@Router			/notes/search [post]
func (h *NoteHandler) Search(w http.ResponseWriter, r *http.Request) {
	user, err := h.deps.JWTAuthentication.GetUser(r)
	if err != nil {
		middleware.Log(r).Error().Err(err).Msg("search note handler unauthorized")
		httpio.Error(w, http.StatusUnauthorized, middleware.ErrUnauthorized)
		return
	}

	request, err := httpio.Parse[search.Request](http.MaxBytesReader(w, r.Body, int64(h.deps.Config.Server.LimitReqJson)))
	if err != nil {
		middleware.Log(r).Debug().Err(err).Msg("search note handler parse request")
		httpio.Error(w, http.StatusBadRequest, err)
		return
	}

	res, err := h.deps.Service.NoteService.Search(r.Context(), user, &request)
	if err != nil {
		if errors.Is(err, note.ErrNoteOperationForbiddenForUser) {
			middleware.Log(r).Error().Err(err).Msg("search note forbidden")
			httpio.Error(w, http.StatusForbidden, errx.New(errx.CodeForbidden, "Operation disallowed"))
			return
		}

		if errors.Is(err, note.ErrNoteSearchBadRequest) {
			middleware.Log(r).Debug().Err(err).Msg("search note bad request")
			httpio.Error(w, http.StatusBadRequest, errx.New(errx.CodeBadRequest, "Bad request"))
			return
		}

		middleware.Log(r).Error().Err(err).Msg("couldn't search notes")
		httpio.Error(w, http.StatusInternalServerError, err)
		return
	}

	httpio.Json(w, http.StatusOK, dtoadapter.NoteSearchToResponseDto(res))
}
