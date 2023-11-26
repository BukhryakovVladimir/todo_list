package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"database/sql"

	"github.com/golang-jwt/jwt/v5"

	_ "github.com/lib/pq"
)

// move to separate file
const secretKey = "chechevitsa"

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
// поменять, чтобы записывал по токену и по айдишнику
func write_taskDB(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	cookie, err := r.Cookie("jwt")

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(404)
		w.Write(resp)
		//defer r.Body.Close()
	} else {
		token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil {
			resp, _ := json.Marshal("unauthenticated")
			w.WriteHeader(403)
			w.Write(resp)
			defer r.Body.Close()
		}

		claims := token.Claims.(*jwt.RegisteredClaims)

		query := `INSERT INTO task(user_id, task_description)
			   	 VALUES($1, $2);`

		bytesBody, errb := io.ReadAll(r.Body)
		if errb != nil {
			panic(errb)
		}

		stringBody := string(bytesBody)

		_, err = db.Exec(query, claims.Issuer, stringBody)

		if err != nil {
			panic(err)
		}

		// var username string

		// if err := db.QueryRow(query, claims.Issuer).Scan(&username); err != nil {
		// 	w.WriteHeader(403)
		// 	w.Write([]byte("username not found"))
		// 	defer r.Body.Close()
		// }
		resp, _ := json.Marshal("Write successful")
		w.WriteHeader(200)
		w.Write(resp)
	}

	//defer r.Body.Close()
}

// удаляет строку таблицы task по номеру строки
func delete_taskDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookie, err := r.Cookie("jwt")

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(404)
		w.Write(resp)
	} else {
		token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		claims := token.Claims.(*jwt.RegisteredClaims)

		query := `DELETE FROM task
		WHERE task_id IN (
		SELECT task_id
		FROM (
		SELECT task_id, user_id, row_number() OVER (PARTITION BY user_id ORDER BY task_id) AS rnum
		FROM task
		)
		WHERE user_id = $1 and rnum = $2
		)`

		bytesBody, err := io.ReadAll(r.Body)

		if err != nil {
			panic(err)
		}

		stringBody := string(bytesBody)

		_, err = db.Exec(query, claims.Issuer, stringBody)

		if err != nil {
			panic(err)
		}
		fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("delete successfully")
		w.WriteHeader(200)
		w.Write(resp)
	}

	// if err != nil {

	// }

	// query := `WITH a AS (SELECT task.*, row_number() OVER () AS rnum FROM task)
	// 		  DELETE FROM task WHERE task_description IN (SELECT task_description FROM a WHERE rnum = $1);`
	// // if task_description is equal it deletes every task_description instead for correct row, this is unacceptable, need to fix
	// //I can do this by adding id to task table, but the serial is not decrementing on delete, and so the id will overflow pretty soon
	// // I am completely mentally stable
	// bytesBody, errb := io.ReadAll(r.Body)

	// if errb != nil {
	// 	panic(errb)
	// }

	// stringBody := string(bytesBody)

	// fmt.Println(stringBody, query)

	// _, err := db.Exec(query, stringBody)

	// if err != nil {
	// 	panic(err)
	// }

}

// Возвращает строки task_descirption из таблицы task
func read_taskDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookie, err := r.Cookie("jwt")

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(404)
		w.Write(resp)
	} else {
		token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		claims := token.Claims.(*jwt.RegisteredClaims)

		query := `SELECT task_description FROM task WHERE user_id = $1;`

		rows, err := db.Query(query, claims.Issuer)
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
	}
}

// Добавляет user_id, username в таблицу user и добавляет user_id, password в таблицу credentials
// Обе операции должны выполнится, поэтому находятся в одной транзакции
// Отношение таблиц	 1 to 1
func signup_userDB(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

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

	//fmt.Println(user.Username, string(password))
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	_, err = tx.ExecContext(ctx, userQuery, user.Username)
	if err != nil {
		tx.Rollback()
		// w.Header()
		// w.Write() smth smth user already exists
		return
	}

	_, err = tx.ExecContext(ctx, credentialsQuery, password)
	if err != nil {
		tx.Rollback()
		// w.Header()
		// w.Write() smth smth user already exists
		return
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

// MEMO: return to task deletion by row problem after doing this
func login_userDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var user User

	bytesBody, err := io.ReadAll(r.Body)

	json.Unmarshal(bytesBody, &user)

	if err != nil {
		panic(err)
	}

	queryUsername := `SELECT user_id 
					 FROM "user"
					 WHERE username = ($1::text)
					 `
	queryPassword := `SELECT password
					  FROM credentials
					  WHERE user_id = $1`

	row := db.QueryRow(queryUsername, user.Username)
	var userId string
	if err := row.Scan(&userId); err != nil {
		resp, _ := json.Marshal("Username not found")
		w.WriteHeader(404)
		w.Write(resp)
		//panic(err) //Do not panic, write response that no user with such login was found instead
	} else {
		//fmt.Println(userId)

		if err := row.Err(); err != nil {
			panic(err)
		}

		row = db.QueryRow(queryPassword, userId)
		var password_hash string
		if err := row.Scan(&password_hash); err != nil {
			resp, _ := json.Marshal("Username not found")
			w.WriteHeader(404)
			w.Write(resp)
		} else {
			//fmt.Println(password_hash)
			if err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(user.Password)); err != nil {
				resp, _ := json.Marshal("Incorrect password")
				w.WriteHeader(401)
				w.Write(resp)
			} else {

				claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Issuer:    userId,
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
				})

				token, err := claims.SignedString([]byte(secretKey))

				if err != nil {
					resp, _ := json.Marshal("Could not login")
					w.WriteHeader(500)
					w.Write(resp)
					defer r.Body.Close()
				}

				tokenCookie := http.Cookie{
					Name:     "jwt",
					Value:    token,
					Expires:  time.Now().Add(time.Hour * 24 * 30),
					HttpOnly: false,
				}

				http.SetCookie(w, &tokenCookie)
				resp, _ := json.Marshal("Successfully loged in")
				w.WriteHeader(200)
				w.Write(resp)
			}
		}
	}
}

func user_userDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cookie, err := r.Cookie("jwt")

	if err != nil {
		resp, _ := json.Marshal("")
		w.WriteHeader(404)
		w.Write(resp)
		defer r.Body.Close()
	} else {
		token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		//fmt.Println(token)
		if err != nil {
			resp, _ := json.Marshal("Unauthenticated")
			w.WriteHeader(403)
			w.Write(resp)
			defer r.Body.Close()
		}
		//fmt.Println(token)
		claims := token.Claims.(*jwt.RegisteredClaims)
		//fmt.Println(claims.Issuer)
		query := `SELECT username FROM "user" WHERE user_id = $1`

		var username string

		if err := db.QueryRow(query, claims.Issuer).Scan(&username); err != nil {
			resp, _ := json.Marshal("Username not found")
			w.WriteHeader(403)
			w.Write(resp)
			defer r.Body.Close()
		} else {
			resp, _ := json.Marshal(username)
			w.WriteHeader(200)
			w.Write(resp)
		}

		//fmt.Println(username)
		//fmt.Println(claims.Issuer)
	}

}

func main() {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*", "http://localhost"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		// /MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	//Go-chi маршрутизатор
	r.Route("/api", func(r chi.Router) {

		r.Post("/write", write_taskDB)

		r.Post("/delete", delete_taskDB)

		r.Post("/signup", signup_userDB)

		r.Get("/read", read_taskDB)

		r.Post("/login", login_userDB)

		r.Get("/user", user_userDB)

	})

	http.ListenAndServe(":3000", r)
}
