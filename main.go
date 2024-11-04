package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Определим структуру User
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

// Мы будем использовать мапу для хранения пользователей
var users = make(map[string]User)
var mu sync.Mutex

// Handler для создания нового пользователя
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	// Прочитаем тело запроса и декодируем JSON в структуру User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// создать уникальный ID (например, просто использовать имя как ID для простоты)
	user.ID = fmt.Sprintf("%d", len(users)+1)

	mu.Lock()             // Защита доступа к мапе пользователей
	users[user.ID] = user // Сохраняем пользователя в мапу
	mu.Unlock()           // Освобождаем блокировку

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(user) // Отправляем назад созданного пользователя
}

// Handler для получения пользователя по ID
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")   // Получаем ID из параметра запроса
	mu.Lock()                 // Защита доступа к мапе пользователей
	user, exists := users[id] // Проверяем, есть ли пользователь с таким ID
	mu.Unlock()               // Освобождаем блокировку

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user) // Отправляем пользователя в формате JSON
}

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/users", createUserHandler)  // Создание пользователя
	m.HandleFunc("/user/{id}", getUserHandler) // Получение пользователя по ID
	http.ListenAndServe(":7777", m)
}
