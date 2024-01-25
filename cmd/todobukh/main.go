package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/handlers/todobukhhandler"

	"github.com/BukhryakovVladimir/todo_list/internal/routes"
)

func main() {

	routes.InitConnPool()

	r := chi.NewRouter()

	todobukhhandler.SetupRoutes(r)

	certFile := "/etc/golang/ssl/localhost.crt" // Следует брать название сертификата из переменных среды (getenv)
	keyFile := "/etc/golang/ssl/localhost.key"  // Следует брать название ключа из переменных среды (getenv)

	// Use ListenAndServeTLS to start the HTTPS server
	err := http.ListenAndServeTLS(":3000", certFile, keyFile, r)
	if err != nil {
		// Handle error
		panic(err)
	}

	// http.ListenAndServe(":3000", r)
}
