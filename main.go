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
	mux.HandleFunc("POST /flags", api.CreateFlag(s))
	mux.HandleFunc("GET /flags", api.ListFlags(s))
	mux.HandleFunc("GET /flags/{key}", api.GetFlag(s))
	mux.HandleFunc("PUT /flags/{key}", api.UpdateFlag(s))
	mux.HandleFunc("DELETE /flags/{key}", api.DeleteFlag(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", api.EvaluateFlag(s))
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
