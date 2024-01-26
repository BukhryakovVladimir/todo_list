package todobukhhandler

import (
	"github.com/go-chi/chi/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/routes"
)

// Устанавливает маршруты
func SetupRoutes(r chi.Router) {
	r.Use(corsMiddleware)

	//Go-chi маршрутизатор
	r.Route("/api", func(r chi.Router) {

		r.Post("/write", routes.Write_taskDB)

		r.Put("/update", routes.Update_taskDB)

		r.Delete("/delete", routes.Delete_taskDB)

		r.Get("/read", routes.Read_taskDB)

		r.Put("/setIsCompletedTrue", routes.SetIsCompletedTrue_taskDB)

		r.Put("/setIsCompletedFalse", routes.SetIsCompletedFalse_taskDB)

		r.Post("/signup", routes.Signup_userDB)

		r.Post("/login", routes.Login_userDB)

		r.Get("/user", routes.User_userDB)

	})
}
