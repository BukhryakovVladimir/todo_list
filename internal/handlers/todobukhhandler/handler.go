package todobukhhandler

import (
	"github.com/go-chi/chi/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/routes"
)

// SetupRoutes устанавливает маршруты
func SetupRoutes(r chi.Router) {
	r.Use(corsMiddleware)

	//Go-chi маршрутизатор
	r.Route("/api", func(r chi.Router) {

		r.Post("/write", routes.WriteTask)

		r.Put("/update", routes.UpdateTask)

		r.Delete("/delete", routes.DeleteTask)

		r.Get("/read", routes.ReadTask)

		r.Put("/setIsCompletedTrue", routes.SetTaskIsCompletedTrue)

		r.Put("/setIsCompletedFalse", routes.SetTaskIsCompletedFalse)

		r.Post("/signup", routes.SignupUser)

		r.Post("/login", routes.LoginUser)

		r.Get("/user", routes.FindUser)

	})
}
