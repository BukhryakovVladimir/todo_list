package todobukhhandler

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Устанавливает политику CORS
func corsMiddleware(next http.Handler) http.Handler {

	strCorsOrigin := os.Getenv("CORS_ORIGIN")
	if strCorsOrigin == "" {
		log.Fatalf("Environment variable CORS_ORIGIN is empty.")
	}
	corsOrigin := fmt.Sprintf("https://%s", strCorsOrigin)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
