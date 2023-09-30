package main

import (
	"context"
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

// База данных
var db *sql.DB

// Структура пользователя
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// создаёт подключение к БД testdb. Выполняеся единожды
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

// записывает task_description в таблицу task
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

	defer r.Body.Close()
}

// удаляет строку таблицы task по номеру строки
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

	defer r.Body.Close()
}

// Возвращает строки task_descirption из таблицы task
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
	defer r.Body.Close()
}

// Добавляет user_id, username в таблицу user и добавляет user_id, password в таблицу credentials
// Обе операции должны выполнится, поэтому находятся в одной транзакции
// Отношение таблиц	 1 to 1
func signup_userDB(w http.ResponseWriter, r *http.Request) {

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

	ctx := context.Background()

	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	_, err = txn.ExecContext(ctx, userQuery, user.Username)
	if err != nil {
		panic(err)
	}

	_, err = txn.ExecContext(ctx, credentialsQuery, password)
	if err != nil {
		panic(err)
	}

	err = txn.Commit()
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()
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

	//Go-chi маршрутизатор
	r.Route("/api", func(r chi.Router) {

		r.Post("/write", write_taskDB)

		r.Post("/delete", delete_taskDB)

		r.Post("/signup", signup_userDB)

		r.Get("/read", read_taskDB)

	})

	http.ListenAndServe(":3000", r)
}
