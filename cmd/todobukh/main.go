package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/handlers/todobukhhandler"

	"github.com/BukhryakovVladimir/todo_list/internal/routes"
)

func main() {

	routes.InitConnPool()

	r := chi.NewRouter()

	todobukhhandler.SetupRoutes(r)

	//TSL_CERT=/etc/golang/ssl/localhost.crt # Полный путь к сертификату
	//TSL_KEY=/etc/golang/ssl/localhost.key # Полный путь к ключу
	//PORT=3000 # Только порт (цеоле число)

	certFile := os.Getenv("TSL_CERT") // Следует брать название сертификата из переменных среды (getenv)
	keyFile := os.Getenv("TSL_KEY")   // Следует брать название ключа из переменных среды (getenv)

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))

	// Use ListenAndServeTLS to start the HTTPS server
	err := http.ListenAndServeTLS(port, certFile, keyFile, r)
	if err != nil {
		// Handle error
		panic(err)
	}

	// http.ListenAndServe(":3000", r)
}
