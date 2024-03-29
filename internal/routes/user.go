package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/BukhryakovVladimir/todo_list/internal/model"
)

// SignupUser добавляет user_id, username в таблицу user и добавляет user_id, password в таблицу credentials.
// Обе операции должны выполнится, поэтому находятся в одной транзакции.
// Отношение таблиц	 1 to 1
func SignupUser(w http.ResponseWriter, r *http.Request) {
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
		resp, err := json.Marshal("Username should have at least 3 characters and consist only of English letters and digits.")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	if !isValidPassword(user.Password) {
		resp, err := json.Marshal("Password should have at least 8 characters and include both English letters and digits. Special characters optionally.")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
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

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v\n", err)
			}
		}
	}()

	var userID int

	// QueryRowContext выполняет первый запрос и возвращает user_id
	err = tx.QueryRowContext(ctx, userQuery, user.Username).Scan(&userID)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("signup_userDB QueryRowContext deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			// Используется вместо http.Error чтобы не было ошибки на фронте.
			resp, err := json.Marshal("Username is taken")
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, err = w.Write(resp)
			if err != nil {
				log.Printf("Write failed: %v\n", err)
			}
			//http.Error(w, "Username is taken", http.StatusConflict) Будет выводить ошибку на фронте. Не менять.
			return
		}
	}

	// ExecContext выполнет второй запрос используя user_id возвращённый первым запросом
	_, err = tx.ExecContext(ctx, credentialsQuery, userID, string(password))
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("signup_userDB ExecContext deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			http.Error(w, "Error inserting credentials", http.StatusInternalServerError)
			return
		}
	}

	// Комит транзакции
	err = tx.Commit()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("signup_userDB Commit deadline exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			http.Error(w, "Error committing transaction", http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal("Signup successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// LoginUser аутентифицирует пользователя.
func LoginUser(w http.ResponseWriter, r *http.Request) {
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

	queryUsername := `SELECT user_id 
					 FROM "user"
					 WHERE username = ($1::text)
					 `
	queryPassword := `SELECT password
					  FROM credentials
					  WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	row := db.QueryRowContext(ctx, queryUsername, user.Username)
	var userId string
	if err := row.Scan(&userId); err != nil {
		resp, err := json.Marshal("Username not found")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	if err := row.Err(); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	row = db.QueryRowContext(ctx, queryPassword, userId)
	var password_hash string
	if err := row.Scan(&password_hash); err != nil {
		resp, err := json.Marshal("Username not found")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(user.Password)); err != nil {
		resp, err := json.Marshal("Incorrect password")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    userId,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
	})

	token, err := claims.SignedString([]byte(secretKey))

	if err != nil {
		resp, err := json.Marshal("Could not login")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	tokenCookie := http.Cookie{
		Name:     jwtName,
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		HttpOnly: false,
	}

	http.SetCookie(w, &tokenCookie)
	resp, err := json.Marshal("Successfully logged in")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// FindUser находит и выводит имя пользователя с помощью jwt cookie.
func FindUser(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Use GET.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, err := json.Marshal("")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}
	token, err := jwtCheck(cookie)

	if err != nil {
		resp, err := json.Marshal("Unauthenticated")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	query := `SELECT username FROM "user" WHERE user_id = $1`

	var username string

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	if err := db.QueryRowContext(ctx, query, claims.Issuer).Scan(&username); err != nil {
		resp, err := json.Marshal("Username not found")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	} else {
		resp, err := json.Marshal(username)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
	}
}

// RegEx. Обязательно латинские буквы, цифры и длина >= 3.
func isValidUsername(username string) bool {
	pattern := "^[a-zA-Z0-9]{3,}$"

	regexpPattern := regexp.MustCompile(pattern)

	return regexpPattern.MatchString(username)
}

// RegEx. Обязательно латинские буквы, цифры и длина >= 8. Опционально специальные символы.
func isValidPassword(password string) bool {
	pattern := `^[a-zA-Z0-9!@#$%^&*()-_=+,.?;:{}|<>]*[a-zA-Z]+[0-9!@#$%^&*()-_=+,.?;:{}|<>]*[0-9]+[a-zA-Z0-9!@#$%^&*()-_=+,.?;:{}|<>]*$`

	regexpPattern := regexp.MustCompile(pattern)

	return regexpPattern.MatchString(password) && len(password) >= 8
}
