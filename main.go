package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"

	"database/sql"

	"github.com/golang-jwt/jwt/v5"

	_ "github.com/lib/pq"
)

// move to separate file
var secretKey string
var jwtName string

// База данных
var db *sql.DB

// Структура пользователя
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Task struct {
	TaskDescription string `json:"task_description"`
	IsCompleted     bool   `json:"is_completed"`
}

// создаёт подключение к БД testdb. Выполняеся единожды
func init() {
	secretKey = os.Getenv("SECRET_KEY")
	jwtName = os.Getenv("JWT_NAME")
	var err error

	// psql := "postgresql://postgres:postgres@todobukh-postgres:5432/todobukh?sslmode=disable"
	// psql := "postgres:postgres@tcp(todobukh-postgres:5432)/todobukh?charset=utf8&parseTime=True&loc=Local"
	// server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	DbHost := os.Getenv("DB_HOST")
	DbPort := os.Getenv("DB_PORT")
	DbUser := os.Getenv("DB_USER")
	DbName := os.Getenv("DB_NAME")
	DbPassword := os.Getenv("DB_PASSWORD")
	psql := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
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
	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(resp)
		//defer r.Body.Close()
	} else {
		token, err := jwtCheck(cookie)

		if err != nil {
			resp, _ := json.Marshal("Unauthenticated")
			w.WriteHeader(http.StatusUnauthorized)
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
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	}
}

// удаляет строку таблицы task по номеру строки.
func delete_taskDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили DELETE request
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid HTTP method. Use DELETE.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(resp)
	} else {
		token, err := jwtCheck(cookie)

		if err != nil {
			resp, _ := json.Marshal("Error while parsing jwt")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(resp)
		}

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
		// fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("delete successfully")
		w.WriteHeader(http.StatusNoContent)
		w.Write(resp)
	}
}

// Возвращает строки task_descirption из таблицы task
func read_taskDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(resp)
	} else {
		token, err := jwtCheck(cookie)

		if err != nil {
			resp, _ := json.Marshal("Error while parsing jwt")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(resp)
		}

		claims := token.Claims.(*jwt.RegisteredClaims)

		query := `SELECT task_description, is_completed FROM task WHERE user_id = $1 ORDER BY task_id;`

		rows, err := db.Query(query, claims.Issuer)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var tasks []Task

		for rows.Next() {
			var task Task
			if err := rows.Scan(&task.TaskDescription, &task.IsCompleted); err != nil {
				log.Fatal(err)
			}

			tasks = append(tasks, task)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		resp, _ := json.Marshal(tasks)

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func setIsCompletedTrue_taskDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили PUT request
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid HTTP method. Use PUT.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(resp)
	} else {
		token, err := jwtCheck(cookie)

		if err != nil {
			resp, _ := json.Marshal("Error while parsing jwt")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(resp)
		}

		claims := token.Claims.(*jwt.RegisteredClaims)

		query := `UPDATE task
		SET is_completed=true
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
		// fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("set is_completed = true successfully")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func setIsCompletedFalse_taskDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили PUT request
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid HTTP method. Use PUT.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, _ := json.Marshal("Unauthenticated")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(resp)
	} else {
		token, err := jwtCheck(cookie)

		if err != nil {
			resp, _ := json.Marshal("Error while parsing jwt")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(resp)
		}

		claims := token.Claims.(*jwt.RegisteredClaims)

		query := `UPDATE task
		SET is_completed=false
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
		// fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("set is_completed=true successfully")
		w.WriteHeader(http.StatusOK)
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
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(bytesBody, &user)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	if !isValidUsername(user.Username) {
		resp, _ := json.Marshal("Username should have at least 3 characters and consist only of English letters and digits.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	if !isValidPassword(user.Password) {
		resp, _ := json.Marshal("Password should have at least 8 characters and include both English letters and digits. Special characters optionally.")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	userQuery := `INSERT INTO "user" (username) VALUES ($1::text) RETURNING user_id;`
	credentialsQuery := `INSERT INTO credentials (user_id, password) VALUES ($1, $2::text);`

	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	var userID int

	// QueryRowContext выполняет первый запрос и возвращает user_id
	err = tx.QueryRowContext(ctx, userQuery, user.Username).Scan(&userID)
	if err != nil {
		tx.Rollback()
		resp, _ := json.Marshal("Username is taken")
		w.WriteHeader(http.StatusConflict)
		w.Write(resp)
		return
	}

	// ExecContext выполнет второй запрос используя user_id возвращённый первым запросом
	_, err = tx.ExecContext(ctx, credentialsQuery, userID, string(password))
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error inserting credentials", http.StatusInternalServerError)
		return
	}

	// Комит транзакции
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	resp, _ := json.Marshal("Signup successful")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
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
		w.WriteHeader(http.StatusNotFound)
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
			w.WriteHeader(http.StatusNotFound)
			w.Write(resp)
		} else {
			//fmt.Println(password_hash)
			if err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(user.Password)); err != nil {
				resp, _ := json.Marshal("Incorrect password")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write(resp)
			} else {

				claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Issuer:    userId,
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
				})

				token, err := claims.SignedString([]byte(secretKey))

				if err != nil {
					resp, _ := json.Marshal("Could not login")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write(resp)
					defer r.Body.Close()
				}

				tokenCookie := http.Cookie{
					Name:     jwtName,
					Value:    token,
					Expires:  time.Now().Add(time.Hour * 24 * 30),
					HttpOnly: false,
				}

				http.SetCookie(w, &tokenCookie)
				resp, _ := json.Marshal("Successfully loged in")
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
			}
		}
	}
}

func user_userDB(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, _ := json.Marshal("")
		w.WriteHeader(http.StatusNotFound)
		w.Write(resp)
		defer r.Body.Close()
	} else {
		token, err := jwtCheck(cookie)

		//fmt.Println(token)
		if err != nil {
			resp, _ := json.Marshal("Unauthenticated")
			w.WriteHeader(http.StatusUnauthorized)
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
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(resp)
			defer r.Body.Close()
		} else {
			resp, _ := json.Marshal(username)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}

		//fmt.Println(username)
		//fmt.Println(claims.Issuer)
	}

}

// jwtCheck парсит JWT токен из переданного HTTP cookie используя секретный ключ secretKey
func jwtCheck(cookie *http.Cookie) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	return token, err
}

// Обязательно латинские буквы, цифры и длина >= 3.
func isValidUsername(username string) bool {
	pattern := "^[a-zA-Z0-9]{3,}$"

	regexpPattern := regexp.MustCompile(pattern)

	return regexpPattern.MatchString(username)
}

// Обязательно латинские буквы, цифры и длина >= 8. Опционально специальные символы.
func isValidPassword(password string) bool {
	pattern := `^[a-zA-Z0-9!@#$%^&*()-_=+,.?;:{}|<>]*[a-zA-Z]+[0-9!@#$%^&*()-_=+,.?;:{}|<>]*[0-9]+[a-zA-Z0-9!@#$%^&*()-_=+,.?;:{}|<>]*$`

	regexpPattern := regexp.MustCompile(pattern)

	return regexpPattern.MatchString(password)
}

func main() {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Access-Control-Allow-Origin", "https://localhost")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	//Go-chi маршрутизатор
	r.Route("/api", func(r chi.Router) {

		r.Post("/write", write_taskDB)

		r.Delete("/delete", delete_taskDB)

		r.Get("/read", read_taskDB)

		r.Put("/setIsCompletedTrue", setIsCompletedTrue_taskDB)

		r.Put("/setIsCompletedFalse", setIsCompletedFalse_taskDB)

		r.Post("/signup", signup_userDB)

		r.Post("/login", login_userDB)

		r.Get("/user", user_userDB)

	})

	certFile := "/etc/golang/ssl/localhost.crt"
	keyFile := "/etc/golang/ssl/localhost.key"

	// Use ListenAndServeTLS to start the HTTPS server
	err := http.ListenAndServeTLS(":3000", certFile, keyFile, r)
	if err != nil {
		// Handle error
		panic(err)
	}

	// http.ListenAndServe(":3000", r)
}
