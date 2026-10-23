package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rinatkh/classroom13/before/internal/note"
)

type NoteUseCase interface {
	Create(ctx context.Context, title, text string) (note.Note, error)
	List(ctx context.Context) ([]note.Note, error)
}

type Handler struct{ useCase NoteUseCase }

func NewHandler(useCase NoteUseCase) *Handler { return &Handler{useCase: useCase} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.useCase.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	result, err := h.useCase.Create(r.Context(), request.Title, request.Text)
	if errors.Is(err, note.ErrEmptyTitle) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
