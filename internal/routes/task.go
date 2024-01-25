package routes

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/BukhryakovVladimir/todo_list/internal/model"
)

const queryTimeLimit int = 5 // заменить на переменные среды

// Записывает task_description в таблицу task
func Write_taskDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method. Use POST.", http.StatusMethodNotAllowed)
		return
	}

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
			defer r.Body.Close() // replace with return, 1 defer at the top of function is enough
		}

		claims := token.Claims.(*jwt.RegisteredClaims)

		query := `INSERT INTO task(user_id, task_description)
			   	 VALUES($1, $2);`

		bytesBody, errb := io.ReadAll(r.Body)
		if errb != nil {
			panic(errb)
		}

		stringBody := string(bytesBody)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
		defer cancel()

		_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Database query time limit exceeded: ", err)
				http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
				return
			} else {
				log.Println("Database query error: ", err)
				http.Error(w, "Database query error", http.StatusInternalServerError)
				return
			}
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

// Обновляет task_description
func Update_taskDB(w http.ResponseWriter, r *http.Request) {

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
		panic(err)
	}

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
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Database query time limit exceeded: ", err)
				http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
				return
			} else {
				log.Println("Database query error: ", err)
				http.Error(w, "Database query error", http.StatusInternalServerError)
				return
			}
		}

		// fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("Task changed successfully")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// Удаляет строку таблицы task по номеру строки.
func Delete_taskDB(w http.ResponseWriter, r *http.Request) {
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
		defer cancel()

		_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Database query time limit exceeded: ", err)
				http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
				return
			} else {
				log.Println("Database query error: ", err)
				http.Error(w, "Database query error", http.StatusInternalServerError)
				return
			}
		}

		// fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("delete successfully")
		w.WriteHeader(http.StatusNoContent)
		w.Write(resp)
	}
}

// Возвращает строки task_descirption из таблицы task
func Read_taskDB(w http.ResponseWriter, r *http.Request) {
	// Используем r.Method чтобы удостовериться что получили GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method. Use GET.", http.StatusMethodNotAllowed)
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

		query := `SELECT task_description, is_completed FROM task WHERE user_id = $1 ORDER BY task_id;`

		rows, err := db.Query(query, claims.Issuer)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var tasks []model.Task

		for rows.Next() {
			var task model.Task
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

// Устанавливает флаг обозначающий выполнение задачи в значение true
func SetIsCompletedTrue_taskDB(w http.ResponseWriter, r *http.Request) {
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
		defer cancel()

		_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Database query time limit exceeded: ", err)
				http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
				return
			} else {
				log.Println("Database query error: ", err)
				http.Error(w, "Database query error", http.StatusInternalServerError)
				return
			}
		}

		// fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("set is_completed = true successfully")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// Устанавливает флаг обозначающий выполнение задачи в значение false
func SetIsCompletedFalse_taskDB(w http.ResponseWriter, r *http.Request) {
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(queryTimeLimit)*time.Second)
		defer cancel()

		_, err = db.ExecContext(ctx, query, claims.Issuer, stringBody)

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Database query time limit exceeded: ", err)
				http.Error(w, "Database query time limit exceeded", http.StatusGatewayTimeout)
				return
			} else {
				log.Println("Database query error: ", err)
				http.Error(w, "Database query error", http.StatusInternalServerError)
				return
			}
		}

		// fmt.Println(claims.Issuer, stringBody)
		resp, _ := json.Marshal("set is_completed=true successfully")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}
