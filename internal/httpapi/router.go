// Package httpapi собирает public и protected HTTP routes.
package httpapi

import (
	"net/http"

	"github.com/rinatkh/classroom13/internal/httpapi/handler"
	"github.com/rinatkh/classroom13/internal/httpapi/middleware"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	noteHandler *handler.NoteHandler,
	authMiddleware *middleware.Auth,
) http.Handler {
	mux := http.NewServeMux()

	// Public endpoints: token ещё нет.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)

	// Protected endpoints: одна и та же authentication-проверка оборачивает
	// каждый handler, но ownership всё равно проверяет Note UseCase.
	protect := authMiddleware.RequireUser
	mux.Handle("GET /notes", protect(http.HandlerFunc(noteHandler.List)))
	mux.Handle("GET /notes/{id}", protect(http.HandlerFunc(noteHandler.Get)))
	mux.Handle("POST /notes", protect(http.HandlerFunc(noteHandler.Create)))
	mux.Handle("PUT /notes/{id}", protect(http.HandlerFunc(noteHandler.Update)))
	mux.Handle("DELETE /notes/{id}", protect(http.HandlerFunc(noteHandler.Delete)))

	return mux
}
