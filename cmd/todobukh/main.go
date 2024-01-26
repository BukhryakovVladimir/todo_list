package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/handlers/todobukhhandler"

	"github.com/BukhryakovVladimir/todo_list/internal/routes"
)

func main() {

	err := routes.InitConnPool()

	if err != nil {
		log.Fatalf("Error connecting to Database: %v", err)
	}

	r := chi.NewRouter()

	todobukhhandler.SetupRoutes(r)

	certFile := os.Getenv("TSL_CERT")
	if certFile == "" {
		log.Fatalf("Environment variable TSL_CERT is empty.")
	}
	keyFile := os.Getenv("TSL_KEY")
	if keyFile == "" {
		log.Fatalf("Environment variable TSL_KEY is empty.")
	}
	strPort := os.Getenv("PORT")
	if strPort == "" {
		log.Fatalf("Environment variable PORT is empty.")
	}
	port := fmt.Sprintf(":%s", strPort)

	// Для запуска HTTPS сервера используется ListenAndServeTLS
	err = http.ListenAndServeTLS(port, certFile, keyFile, r)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
