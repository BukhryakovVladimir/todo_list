package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"database/sql"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error

	psql := "postgresql://postgres:postgres@localhost:2022/postgres?sslmode=disable"

	db, err = sql.Open("postgres", psql)

	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("connecte to postgree")
}

func main() {
	r := chi.NewRouter()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {

	})

	http.ListenAndServe(":3000", r)
}
