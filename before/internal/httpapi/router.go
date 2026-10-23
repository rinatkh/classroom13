package httpapi

import "net/http"

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /notes", handler.List)
	mux.HandleFunc("POST /notes", handler.Create)
	return mux
}
