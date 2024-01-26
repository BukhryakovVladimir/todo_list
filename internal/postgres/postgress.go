package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// создаёт подключение к БД
func Dial() (*sql.DB, error) {

	var db *sql.DB // Пул соединений с БД

	DbHost := os.Getenv("DB_HOST")
	if DbHost == "" {
		return nil, errors.New("Environment variable DB_HOST is empty.")
	}
	DbPort := os.Getenv("DB_PORT")
	if DbPort == "" {
		return nil, errors.New("Environment variable DB_PORT is empty.")
	}
	DbUser := os.Getenv("DB_USER")
	if DbUser == "" {
		return nil, errors.New("Environment variable DB_USER is empty.")
	}
	DbName := os.Getenv("DB_NAME")
	if DbName == "" {
		return nil, errors.New("Environment variable DB_NAME is empty.")
	}
	DbPassword := os.Getenv("DB_PASSWORD")
	if DbPassword == "" {
		return nil, errors.New("Environment variable DB_PASSWORD is empty.")
	}
	strMaxOpenConns := os.Getenv("MAX_OPEN_CONNS")
	if strMaxOpenConns == "" {
		return nil, errors.New("Environment variable MAX_OPEN_CONNS is empty.")
	}
	strMaxIdleConns := os.Getenv("MAX_IDLE_CONNS")
	if strMaxIdleConns == "" {
		return nil, errors.New("Environment variable MAX_IDLE_CONNS is empty.")
	}
	strConnMaxLifetime := os.Getenv("CONN_MAX_LIFETIME")
	if strConnMaxLifetime == "" {
		return nil, errors.New("Environment variable CONN_MAX_LIFETIME is empty.")
	}
	maxOpenConns, err := strconv.Atoi(strMaxOpenConns)
	if err != nil {
		return nil, err
	}
	maxIdleConns, err := strconv.Atoi(strMaxIdleConns)
	if err != nil {
		return nil, err
	}
	connMaxLifetime, err := strconv.Atoi(strConnMaxLifetime)
	if err != nil {
		return nil, err
	}

	psql := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	db, err = sql.Open("postgres", psql)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)

	db.SetMaxIdleConns(maxIdleConns)

	db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatalf("Error starting server: %v", err)
		return nil, err
	}

	fmt.Println("connected to postgree")

	return db, nil
}
