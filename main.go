package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"database/sql"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error

	psql := "postgresql://postgres:postgres@localhost:2022/testdb?sslmode=disable"

	db, err = sql.Open("postgres", psql)

	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("connected to postgree")
}

func writeDB(w http.ResponseWriter, r *http.Request) {
	query := `INSERT INTO task(task_description)
			   VALUES($1);`

	bytesBody, errb := io.ReadAll(r.Body)
	if errb != nil {
		panic(errb)
	}

	stringBody := string(bytesBody)

	_, err := db.Exec(query, stringBody)

	if err != nil {
		panic(err)
	}
}

func deleteDB(w http.ResponseWriter, r *http.Request) {
	query := `WITH a AS (SELECT task.*, row_number() OVER () AS rnum FROM task)
			  DELETE FROM task WHERE task_description IN (SELECT task_description FROM a WHERE rnum = $1)`

	bytesBody, errb := io.ReadAll(r.Body)

	if errb != nil {
		panic(errb)
	}

	stringBody := string(bytesBody)

	fmt.Println(stringBody, query)

	_, err := db.Exec(query, stringBody)

	if err != nil {
		panic(err)
	}
}

func readDB(w http.ResponseWriter, r *http.Request) {
	query := `SELECT task_description FROM task;`

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	tasks := make([]string, 0)
	for rows.Next() {
		var task_description string
		if err := rows.Scan(&task_description); err != nil {
			log.Fatal(err)
		}
		//w.WriteHeader()
		tasks = append(tasks, task_description)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	resp, err := json.Marshal(tasks)

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	w.Write(resp)
}

func main() {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*", "null"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/api", func(r chi.Router) {

		r.Post("/write", writeDB)

		r.Post("/delete", deleteDB)

		r.Get("/read", readDB)

	})

	http.ListenAndServe(":3000", r)
}
