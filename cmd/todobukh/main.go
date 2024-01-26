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
	keyFile := os.Getenv("TSL_KEY")

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))

	// Для запуска HTTPS сервера используется ListenAndServeTLS
	err = http.ListenAndServeTLS(port, certFile, keyFile, r)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
