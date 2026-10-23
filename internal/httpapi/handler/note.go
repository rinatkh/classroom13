package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rinatkh/classroom13/internal/httpapi/authctx"
	"github.com/rinatkh/classroom13/internal/httpapi/dto"
	"github.com/rinatkh/classroom13/internal/httpapi/mapper"
	"github.com/rinatkh/classroom13/internal/note"
)

type NoteUseCase interface {
	Create(ctx context.Context, userID int64, input note.CreateInput) (note.Note, error)
	List(ctx context.Context, userID int64) ([]note.Note, error)
	Get(ctx context.Context, userID, noteID int64) (note.Note, error)
	Update(ctx context.Context, userID, noteID int64, input note.UpdateInput) (note.Note, error)
	Delete(ctx context.Context, userID, noteID int64) error
}

type NoteHandler struct {
	useCase NoteUseCase
}

func NewNoteHandler(useCase NoteUseCase) *NoteHandler {
	return &NoteHandler{useCase: useCase}
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	result, err := h.useCase.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, mapper.NotesToResponse(result))
}

func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	noteID, ok := parseID(w, r)
	if !ok {
		return
	}
	result, err := h.useCase.Get(r.Context(), userID, noteID)
	if !writeNoteError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, mapper.NoteToResponse(result))
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var request dto.CreateNoteRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	result, err := h.useCase.Create(r.Context(), userID, note.CreateInput{
		Title: request.Title, Text: request.Text,
	})
	if !writeNoteError(w, err) {
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/notes/%d", result.ID))
	writeJSON(w, http.StatusCreated, mapper.NoteToResponse(result))
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	noteID, ok := parseID(w, r)
	if !ok {
		return
	}
	var request dto.UpdateNoteRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	result, err := h.useCase.Update(r.Context(), userID, noteID, note.UpdateInput{
		Title: request.Title, Text: request.Text,
	})
	if !writeNoteError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, mapper.NoteToResponse(result))
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	noteID, ok := parseID(w, r)
	if !ok {
		return
	}
	if !writeNoteError(w, h.useCase.Delete(r.Context(), userID, noteID)) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func requireUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return 0, false
	}
	return userID, true
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return 0, false
	}
	return id, true
}

// writeNoteError возвращает true, если ошибки нет и handler может продолжить.
func writeNoteError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, note.ErrEmptyTitle):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, note.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, note.ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
	return false
}
