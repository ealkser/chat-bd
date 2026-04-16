// Package main — это точка входа в веб-приложение для управления аутентификацией пользователей.
//
// Приложение предоставляет API для:
//   - Регистрации (/register)
//   - Входа (/login)
//   - Получения профиля пользователя (/me)
//   - Выхода (/logout)
//   - Обновления токена (/refresh)
//
// Сервер запускается на порту :8080 и использует SQLite для хранения данных,
// JWT для аутентификации и stateless сессий.
package main

import (
	"chat-bd/db"
	"chat-bd/handlers"
	"log"
	"net/http"
)

// main — основная функция приложения.
// Инициализирует базу данных, регистрирует HTTP-обработчики и запускает сервер.
//
// Процесс запуска:
//  1. Вызывается db.InitDB() — подключение к БД и создание таблиц (если нужно).
//  2. Регистрируются маршруты через http.HandleFunc.
//  3. Сервер начинает прослушивание на http://localhost:8080.
//
// При ошибках запуска сервера (например, порт занят), приложение завершается с логом.
func main() {
	// Инициализация подключения к базе данных и создание необходимых таблиц
	db.InitDB()

	// Регистрация маршрутов API
	http.HandleFunc("/register", handlers.RegisterHandler) // POST: регистрация нового пользователя
	http.HandleFunc("/login", handlers.LoginHandler)       // POST: аутентификация и выдача токенов
	http.HandleFunc("/me", handlers.MeHandler)             // GET: получение данных текущего пользователя
	http.HandleFunc("/logout", handlers.LogoutHandler)     // POST: выход (удаление refresh token)
	http.HandleFunc("/refresh", handlers.RefreshHandler)   // POST: обновление access token

	handler := corsMiddleware(http.DefaultServeMux)

	// Запуск HTTP-сервера
	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// corsMiddleware добавляет заголовки CORS, разрешая запросы с фронтенда
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешить запросы с любого origin (для разработки)
		// В продакшене укажите конкретный домен, например: "http://localhost:3000"
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Разрешить определённые заголовки
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Refresh-Token")

		// Разрешить определённые методы
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		// Поддержка credentials (если будет использоваться Access-Control-Allow-Origin с конкретным доменом)
		// w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Обработка preflight-запросов (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передача управления следующему обработчику
		next.ServeHTTP(w, r)
	})
}
