package routes

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/BukhryakovVladimir/todo_list/internal/model"
)

// Добавляет user_id, username в таблицу user и добавляет user_id, password в таблицу credentials
// Обе операции должны выполнится, поэтому находятся в одной транзакции
// Отношение таблиц	 1 to 1
func Signup_userDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var user model.User

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	var userID int

	// QueryRowContext выполняет первый запрос и возвращает user_id
	err = tx.QueryRowContext(ctx, userQuery, user.Username).Scan(&userID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			tx.Rollback()
			log.Println("signup_userDB QueryRowContext deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
		} else {
			tx.Rollback()
			// Added instead of http.Error to avoid Not a valid json error on frontend
			resp, _ := json.Marshal("Username is taken")
			w.WriteHeader(http.StatusConflict)
			w.Write(resp)
			//http.Error(w, "Username is taken", http.StatusConflict)    Not a valid json on frontend error
			return
		}
	}

	// ExecContext выполнет второй запрос используя user_id возвращённый первым запросом
	_, err = tx.ExecContext(ctx, credentialsQuery, userID, string(password))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			tx.Rollback()
			log.Println("signup_userDB ExecContext deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
		} else {
			tx.Rollback()
			http.Error(w, "Error inserting credentials", http.StatusInternalServerError)
			return
		}
	}

	// Комит транзакции
	err = tx.Commit()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			tx.Rollback()
			log.Println("signup_userDB Commit deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Error committing transaction", http.StatusInternalServerError)
			return
		}
	}

	resp, _ := json.Marshal("Signup successful")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// MEMO: return to task deletion by row problem after doing this
func Login_userDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	var user model.User

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

func User_userDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Use GET.", http.StatusMethodNotAllowed)
		return
	}

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
