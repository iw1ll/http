package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

// Определим структуру User
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type UserResponse struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

var db *sql.DB

// InitDB - функция запускает соединение с базой данных и создает таблицу пользователей
func InitDB() {
	var err error
	connStr := "user=postgres password=changeme dbname=postgres sslmode=disable host=localhost port=5432"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем таблицу пользователей
	_, err = db.Exec(`
	 CREATE TABLE IF NOT EXISTS users (
	  id SERIAL PRIMARY KEY,
	  name TEXT,
	  age INT,
	  email TEXT
	 )
	`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database initialized!")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User

	// Прочитаем тело запроса и декодируем JSON в структуру User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Вставляем нового пользователя в базу данных
	err := db.QueryRow(`INSERT INTO users (name, age, email) VALUES ($1, $2, $3) RETURNING id`, user.Name, user.Age, user.Email).Scan(&user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user) // Отправляем назад созданного пользователя
}

func usersAll(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT * FROM users`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users) // Отправляем всех пользователей
}

func getUserId(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // Извлекаем ID из параметра запроса
	var user User

	err := db.QueryRow(`SELECT * FROM users WHERE id = $1`, id).Scan(&user.ID, &user.Name, &user.Age, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user) // Отправляем пользователя в формате JSON
}

func updateUserId(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var user User

	// Прочитаем тело запроса и декодируем JSON в структуру User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Обновляем пользователя в базе данных
	_, err := db.Exec(`UPDATE users SET name = $1, age = $2, email = $3 WHERE id = $4`, user.Name, user.Age, user.Email, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userResponse := UserResponse{
		Name:  user.Name,
		Age:   user.Age,
		Email: user.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponse) // Отправляем пользователя в формате JSON
}

func deleteUserId(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Удаляем пользователя из базы данных
	_, err := db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Возвращаем статус 204 No Content
}

func main() {
	InitDB()
	defer db.Close()
	m := http.NewServeMux()
	m.HandleFunc("POST /create", createUser)
	m.HandleFunc("GET /user/{id}", getUserId)
	m.HandleFunc("GET /users/all", usersAll)
	m.HandleFunc("PUT /users/update/{id}/", updateUserId)
	m.HandleFunc("DELETE /users/delete/{id}/", deleteUserId)

	http.ListenAndServe(":7777", m)
}
