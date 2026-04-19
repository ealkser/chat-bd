// main.go
package main

import (
	"chat-bd/db"
	"chat-bd/handlers"
	"chat-bd/repositories"
	"chat-bd/services"
	"chat-bd/ws"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	// Инициализация БД
	db.InitDB()
	defer db.DB.Close()

	// Создаём репозиторий и сервис чатов
	chatRepo := repositories.NewChatRepository(db.DB)
	chatService := services.NewChatService(chatRepo)

	hub := ws.NewHub(chatService)
	go hub.Run()

	handlers.Init(chatService, hub)

	// Создаём chi-роутер
	r := chi.NewRouter()

	// CORS Middleware (лучше использовать встроенный от chi)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // Для разработки
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // seconds
	}))

	// Маршруты аутентификации
	r.Post("/api/auth/register", handlers.RegisterHandler)
	r.Post("/api/auth/login", handlers.LoginHandler)
	r.Post("/api/auth/logout", handlers.LogoutHandler)
	r.Post("/api/auth/refresh", handlers.RefreshHandler)
	r.Post("/api/auth/verify-code", handlers.VerifyCodeHandler)

	r.Get("/api/me", handlers.MeHandler)

	// Маршруты чатов
	r.Get("/api/chats", handlers.GetChatsHandler)
	r.Get("/api/chat/{id}/messages", handlers.GetMessagesHandler)
	r.Post("/api/chat/{id}/message", handlers.SendMessageHandler)
	r.Post("/api/start-chat", handlers.StartChatHandler)
	r.Get("/api/ws", hub.ServeWS)

	// Запуск сервера
	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
