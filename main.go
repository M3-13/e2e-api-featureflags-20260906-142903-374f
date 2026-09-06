package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"featureflags/internal/api"
	"featureflags/internal/store"
)

func newHandler() http.Handler {
	s := store.NewStore()

	mux := http.NewServeMux()
	mux.Handle("POST /flags", api.RequireAuth(api.CreateFlag(s)))
	mux.HandleFunc("GET /flags", api.ListFlags(s))
	mux.HandleFunc("GET /flags/{key}", api.GetFlag(s))
	mux.Handle("PUT /flags/{key}", api.RequireAuth(api.UpdateFlag(s)))
	mux.Handle("DELETE /flags/{key}", api.RequireAuth(api.DeleteFlag(s)))
	mux.HandleFunc("GET /flags/{key}/evaluate", api.EvaluateFlag(s))
	mux.HandleFunc("GET /healthz", api.Healthz)

	return api.Logging(mux)
}

func main() {
	handler := newHandler()

	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := host + ":" + port
	log.Printf("listening on %s", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
