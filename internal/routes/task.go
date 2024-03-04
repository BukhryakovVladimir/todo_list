package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/model"
)

// WriteTask записывает task_description в таблицу task
func WriteTask(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, err := json.Marshal("Unauthenticated")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
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
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	query := `INSERT INTO task(user_id, task_description)
			   	 VALUES($1, $2);`

	bytesBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stringBody := string(bytesBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("Database query time limit exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			log.Println("Database query error: ", err)
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal("Write successful")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// UpdateTask обновляет task_description
func UpdateTask(w http.ResponseWriter, r *http.Request) {

	// Используем r.Method чтобы удостовериться что получили PUT request
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid HTTP method. Use PUT.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var updateTask model.UpdateTask

	bytesBody, err := io.ReadAll(r.Body)

	json.Unmarshal(bytesBody, &updateTask)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	cookie, err := r.Cookie(jwtName)
	if err != nil {
		resp, err := json.Marshal("Unauthenticated")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	token, err := jwtCheck(cookie)
	if err != nil {
		resp, err := json.Marshal("Error while parsing jwt")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	query := `UPDATE task
		SET task_description=$1
		WHERE task_id IN (
		SELECT task_id
		FROM (
		SELECT task_id, user_id, row_number() OVER (PARTITION BY user_id ORDER BY task_id) AS rnum
		FROM task
		)
		WHERE user_id = $2 and rnum = $3
		)`

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, updateTask.TaskDescription, claims.Issuer, updateTask.TaskNumber)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("Database query time limit exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			log.Println("Database query error: ", err)
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal("Task changed successfully")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// DeleteTask удаляет строку таблицы task по номеру строки.
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили DELETE request
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid HTTP method. Use DELETE.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, err := json.Marshal("Unauthenticated")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}
	token, err := jwtCheck(cookie)

	if err != nil {
		resp, err := json.Marshal("Error while parsing jwt")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stringBody := string(bytesBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("Database query time limit exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			log.Println("Database query error: ", err)
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal("delete successfully")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// ReadTask возвращает строки task_description из таблицы task
func ReadTask(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Use GET.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, err := json.Marshal("Unauthenticated")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}
	token, err := jwtCheck(cookie)

	if err != nil {
		resp, err := json.Marshal("Error while parsing jwt")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	query := `SELECT task_description, is_completed FROM task WHERE user_id = $1 ORDER BY task_id;`

	rows, err := db.Query(query, claims.Issuer)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.TaskDescription, &task.IsCompleted); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// SetTaskIsCompletedTrue устанавливает флаг обозначающий выполнение задачи в значение true
func SetTaskIsCompletedTrue(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили PUT request
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid HTTP method. Use PUT.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, err := json.Marshal("Unauthenticated")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}
	token, err := jwtCheck(cookie)

	if err != nil {
		resp, err := json.Marshal("Error while parsing jwt")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stringBody := string(bytesBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("Database query time limit exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			log.Println("Database query error: ", err)
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal("set is_completed = true successfully")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

// SetTaskIsCompletedFalse устанавливает флаг обозначающий выполнение задачи в значение false
func SetTaskIsCompletedFalse(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили PUT request
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid HTTP method. Use PUT.", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	cookie, err := r.Cookie(jwtName)

	if err != nil {
		resp, err := json.Marshal("Unauthenticated")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
	}
	token, err := jwtCheck(cookie)

	if err != nil {
		resp, err := json.Marshal("Error while parsing jwt")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, err = w.Write(resp)
		if err != nil {
			log.Printf("Write failed: %v\n", err)
		}
		return
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stringBody := string(bytesBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Println("Database query time limit exceeded: ", err)
			http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
			return
		} else {
			log.Println("Database query error: ", err)
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
	}

	resp, err := json.Marshal("set is_completed=true successfully")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}
