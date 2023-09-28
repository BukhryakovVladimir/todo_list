package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"database/sql"

	_ "github.com/lib/pq"
)

var db *sql.DB

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

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

func write_taskDB(w http.ResponseWriter, r *http.Request) {
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

func delete_taskDB(w http.ResponseWriter, r *http.Request) {
	query := `WITH a AS (SELECT task.*, row_number() OVER () AS rnum FROM task)
			  DELETE FROM task WHERE task_description IN (SELECT task_description FROM a WHERE rnum = $1);`

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

func read_taskDB(w http.ResponseWriter, r *http.Request) {
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

func signup_userDB(w http.ResponseWriter, r *http.Request) {
	// bytesBody, errb := io.ReadAll(r.Body)
	// if errb != nil {
	// 	panic(errb)
	// }
	var user User

	bytesBody, err := io.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	json.Unmarshal(bytesBody, &user)

	userQuery := `INSERT INTO "user" (username)
				  VALUES ($1::text);`

	credentialsQuery := `INSERT INTO credentials (password)
						 VALUES($1::text);`
	password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)

	fmt.Println(user.Username, string(password))

	_, errUser := db.Exec(userQuery, user.Username)
	if errUser != nil {
		fmt.Println("1")
		panic(errUser)
	}

	_, errCredentials := db.Exec(credentialsQuery, string(password))
	if errCredentials != nil {
		fmt.Println("2")
		panic(errCredentials)
	}
	fmt.Println(r.Body, "\n", user)
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

		r.Post("/write", write_taskDB)

		r.Post("/delete", delete_taskDB)

		r.Post("/signup", signup_userDB)

		r.Get("/read", read_taskDB)

	})

	http.ListenAndServe(":3000", r)
}
