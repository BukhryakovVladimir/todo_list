package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const (
	MaxOpenConns    int = 10 // Можно заменить на переменные среды
	MaxIdleConns    int = 5  // Можно заменить на переменные среды
	ConnMaxLifetime int = 30 // Можно заменить на переменные среды
)

// создаёт подключение к БД
func Dial() (*sql.DB, error) {

	var db *sql.DB // Пул соединений с БД

	var err error

	DbHost := os.Getenv("DB_HOST")
	DbPort := os.Getenv("DB_PORT")
	DbUser := os.Getenv("DB_USER")
	DbName := os.Getenv("DB_NAME")
	DbPassword := os.Getenv("DB_PASSWORD")

	psql := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	db, err = sql.Open("postgres", psql)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
		return nil, err
	}

	db.SetMaxOpenConns(10)

	db.SetMaxIdleConns(5)

	db.SetConnMaxLifetime(30 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatalf("Error starting server: %v", err)
		return nil, err
	}

	fmt.Println("connected to postgree")

	return db, nil
}
