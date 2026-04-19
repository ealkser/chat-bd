// handlers/chats.go
package handlers

import (
	"chat-bd/models"
	"chat-bd/services"
	"chat-bd/utils"
	"chat-bd/ws"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	chatService *services.ChatService
	Hub         *ws.Hub // ← добавьте
)

func Init(service *services.ChatService, hub *ws.Hub) {
	chatService = service
	Hub = hub // ← сохраняем
}

func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	userID, err := utils.GetUserIDFromToken(authHeader)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неавторизован"}, http.StatusUnauthorized)
		return
	}

	chats, err := chatService.GetChatsForUser(userID)
	if err != nil {
		log.Printf("Ошибка получения чатов: %v", err)
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка сервера"}, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, utils.Response{Success: true, Data: chats}, http.StatusOK)
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неверный ID чата"}, http.StatusBadRequest)
		return
	}
	authHeader := r.Header.Get("Authorization")
	userID, err := utils.GetUserIDFromToken(authHeader)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неавторизован"}, http.StatusUnauthorized)
		return
	}

	inChat, err := chatService.IsUserInChat(chatID, userID)
	if err != nil || !inChat {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Нет доступа к чату"}, http.StatusForbidden)
		return
	}

	messages, err := chatService.GetMessages(chatID)
	if err != nil {
		log.Printf("Ошибка получения сообщений: %v", err)
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка сервера"}, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, utils.Response{Success: true, Data: messages}, http.StatusOK)
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Извлекаем chatID из URL
	chatIDStr := chi.URLParam(r, "id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неверный ID чата"}, http.StatusBadRequest)
		return
	}

	// 2. Получаем userID из токена
	authHeader := r.Header.Get("Authorization")
	userID, err := utils.GetUserIDFromToken(authHeader)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неавторизован"}, http.StatusUnauthorized)
		return
	}

	// 3. Читаем тело запроса
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Некорректные данные"}, http.StatusBadRequest)
		return
	}

	// 4. Отправляем сообщение в БД
	messageID, err := chatService.SendMessage(chatID, userID, req.Content)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка отправки"}, http.StatusInternalServerError)
		return
	}

	// ✅ 5. Формируем сообщение для WebSocket
	message := models.Message{
		ID:        messageID,
		ChatID:    chatID,
		SenderID:  userID,
		Content:   req.Content,
		Read:      false,
		CreatedAt: time.Now().Unix(),
	}

	// ✅ 6. Отправляем всем через WebSocket
	// Hub.BroadcastToRoom(chatID, ws.MessagePayload{
	// 	Event: "new_message",
	// 	Data:  message,
	// })
	data, _ := json.Marshal(message)
	Hub.SendToRoom(chatID, data, message.SenderID)

	// 7. Отвечаем клиенту
	utils.RespondJSON(w, utils.Response{
		Success: true,
		Data:    map[string]int64{"message_id": messageID},
	}, http.StatusOK)
}

func StartChatHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	userID, err := utils.GetUserIDFromToken(authHeader)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неавторизован"}, http.StatusUnauthorized)
		return
	}

	var req struct {
		WithEmail string `json:"with_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Некорректные данные"}, http.StatusBadRequest)
		return
	}

	targetID, exists, err := chatService.UserExists(req.WithEmail)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка поиска пользователя"}, http.StatusInternalServerError)
		return
	}
	if !exists {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Пользователь не найден"}, http.StatusNotFound)
		return
	}

	chatID, err := chatService.StartChat(userID, targetID)
	if err != nil {
		log.Printf("Ошибка старта чата: %v", err)
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка создания чата"}, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, utils.Response{
		Success: true,
		Data:    map[string]int64{"chat_id": chatID},
	}, http.StatusOK)
}
