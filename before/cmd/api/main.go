package main

import (
	"log"
	"net/http"
	"time"

	"github.com/rinatkh/classroom13/before/internal/httpapi"
	"github.com/rinatkh/classroom13/before/internal/note"
	"github.com/rinatkh/classroom13/before/internal/repository/memory"
)

func main() {
	repository := memory.NewNoteRepository()
	useCase := note.NewUseCase(repository)
	handler := httpapi.NewHandler(useCase)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewRouter(handler),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
