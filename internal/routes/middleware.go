package routes

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/postgres"

	_ "github.com/lib/pq"
)

var db *sql.DB // Пул соединений с БД

var (
	secretKey string
	jwtName   string
)

// jwtCheck парсит JWT токен из переданного HTTP cookie используя секретный ключ secretKey
func jwtCheck(cookie *http.Cookie) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	return token, err
}

func InitConnPool() error {
	var err error
	secretKey = os.Getenv("SECRET_KEY")
	jwtName = os.Getenv("JWT_NAME")

	db, err = postgres.Dial()
	if err != nil {
		return err
	}
	return nil
}
