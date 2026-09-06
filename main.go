package main

import (
	"log"
	"net/http"
	"os"

	"featureflags/internal/api"
	"featureflags/internal/store"
)

func main() {
	s := store.NewStore()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", api.Create(s))
	mux.HandleFunc("GET /flags", api.List(s))
	mux.HandleFunc("GET /flags/{key}", api.Get(s))
	mux.HandleFunc("PUT /flags/{key}", api.Update(s))
	mux.HandleFunc("DELETE /flags/{key}", api.Delete(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", api.Evaluate(s))
	mux.HandleFunc("GET /healthz", api.Healthz)

	handler := api.Logging(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
