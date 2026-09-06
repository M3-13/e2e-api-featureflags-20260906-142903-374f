package api

import (
	"log"
	"net/http"
	"os"
)

var accessLog = log.New(os.Stdout, "", 0)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		accessLog.Printf("%s %s %d", r.Method, r.URL.Path, rec.status)
	})
}
