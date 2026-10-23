package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rinatkh/classroom13/internal/auth"
	"github.com/rinatkh/classroom13/internal/httpapi"
	"github.com/rinatkh/classroom13/internal/httpapi/handler"
	"github.com/rinatkh/classroom13/internal/httpapi/middleware"
	"github.com/rinatkh/classroom13/internal/note"
	"github.com/rinatkh/classroom13/internal/repository/memory"
	"github.com/rinatkh/classroom13/internal/user"
)

func main() {
	addr := envOr("HTTP_ADDR", ":8080")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	ttl, err := time.ParseDuration(envOr("TOKEN_TTL", "2h"))
	if err != nil {
		log.Fatalf("parse TOKEN_TTL: %v", err)
	}

	// Composition root: concrete implementations встречаются здесь и
	// передаются в слои через constructors.
	userRepository := memory.NewUserRepository()
	noteRepository := memory.NewNoteRepository()
	passwordHasher := auth.NewBcryptHasher(0)
	tokenManager := auth.NewJWTManager(secret, ttl)

	authUseCase := user.NewUseCase(userRepository, passwordHasher, tokenManager)
	noteUseCase := note.NewUseCase(noteRepository)

	authHandler := handler.NewAuthHandler(authUseCase)
	noteHandler := handler.NewNoteHandler(noteUseCase)
	authMiddleware := middleware.NewAuth(tokenManager)
	router := httpapi.NewRouter(authHandler, noteHandler, authMiddleware)

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Notes API listens on %s", addr)
	log.Fatal(server.ListenAndServe())
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
