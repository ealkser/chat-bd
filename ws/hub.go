// package ws

// import (
// 	"chat-bd/models"
// 	"chat-bd/services"
// 	"chat-bd/utils"
// 	"encoding/json"
// 	"log"
// 	"net/http"

// 	"github.com/gorilla/websocket"
// )

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool { return true },
// }

// type Hub struct {
// 	clients    map[*Client]bool
// 	rooms      map[int64]map[*Client]bool
// 	broadcast  chan MessagePayload
// 	register   chan *Client
// 	unregister chan *Client
// 	service    *services.ChatService // ← добавляем
// }

// type Client struct {
// 	hub    *Hub
// 	conn   *websocket.Conn
// 	send   chan []byte
// 	userID int64
// }

// type MessagePayload struct {
// 	Event string         `json:"event"`
// 	Data  models.Message `json:"data"`
// }

// func NewHub(service *services.ChatService) *Hub {
// 	return &Hub{
// 		clients:    make(map[*Client]bool),
// 		rooms:      make(map[int64]map[*Client]bool),
// 		broadcast:  make(chan MessagePayload),
// 		register:   make(chan *Client),
// 		unregister: make(chan *Client),
// 		service:    service, // ← сохраняем
// 	}
// }

// func (h *Hub) Run() {
// 	for {
// 		select {
// 		case client := <-h.register:
// 			h.clients[client] = true

// 		case client := <-h.unregister:
// 			// Удалить из clients
// 			if _, ok := h.clients[client]; ok {
// 				delete(h.clients, client)
// 			}

// 			// Удалить из всех rooms
// 			for chatID, clients := range h.rooms {
// 				if _, ok := clients[client]; ok {
// 					delete(clients, client)
// 					if len(clients) == 0 {
// 						delete(h.rooms, chatID)
// 					}
// 				}
// 			}

// 			// Закрыть канал ОТПРАВКИ
// 			close(client.send) // ✅ Только здесь!

// 		case message := <-h.broadcast:
// 			for client := range h.clients {
// 				if client.userID == message.Data.SenderID {
// 					continue
// 				}
// 				select {
// 				case client.send <- message.ToJSON():
// 				default:
// 					h.unregister <- client // ✅ Правильно
// 				}
// 			}
// 		}
// 	}
// }

// func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("Ошибка WebSocket: %v", err)
// 		return
// 	}
// 	authHeader := r.URL.Query().Get("token")

// 	// authHeader := r.Header.Get("Authorization")
// 	userID, err := utils.GetUserIDFromToken("Bearer " + authHeader)
// 	if err != nil {
// 		log.Printf("Ошибка WebSocket: %v", err)
// 		conn.Close()
// 		return
// 	}

// 	client := &Client{hub: h, conn: conn, send: make(chan []byte), userID: userID}
// 	h.register <- client

// 	chatIDs, err := h.service.GetChatIDsForUser(userID)
// 	if err != nil {
// 		log.Printf("Ошибка получения чатов для пользователя %d: %v", userID, err)
// 		// Можно закрыть соединение или продолжить без подписки
// 		chatIDs = []int64{} // пусто
// 	}

// 	for _, chatID := range chatIDs {
// 		if _, ok := h.rooms[chatID]; !ok {
// 			h.rooms[chatID] = make(map[*Client]bool)
// 		}
// 		h.rooms[chatID][client] = true
// 	}

// 	go client.writePump()
// 	go client.readPump()
// }

// func (h *Hub) Broadcast(payload MessagePayload) {
// 	h.broadcast <- payload
// }

// func (h *Hub) BroadcastToRoom(chatID int64, payload MessagePayload) {
// 	clients, ok := h.rooms[chatID]
// 	if !ok {
// 		return
// 	}
// 	for client := range clients {
// 		if client.userID == payload.Data.SenderID {
// 			continue
// 		}
// 		select {
// 		case client.send <- payload.ToJSON():
// 		default:
// 			h.unregister <- client // ✅ Только это
// 		}
// 	}
// }

// func (c *Client) readPump() {
// 	defer func() {
// 		c.hub.unregister <- c
// 		c.conn.Close()
// 	}()

// 	for {
// 		_, message, err := c.conn.ReadMessage()
// 		if err != nil {
// 			break
// 		}

// 		var req map[string]interface{}
// 		if json.Unmarshal(message, &req) == nil {
// 			if event, ok := req["event"]; ok {
// 				switch event {
// 				case "join":
// 					if chatID, ok := req["chat_id"].(float64); ok {
// 						chatIDInt := int64(chatID)
// 						if _, ok := c.hub.rooms[chatIDInt]; !ok {
// 							c.hub.rooms[chatIDInt] = make(map[*Client]bool)
// 						}
// 						c.hub.rooms[chatIDInt][c] = true
// 					}

// 				case "leave":
// 					if chatID, ok := req["chat_id"].(float64); ok {
// 						chatIDInt := int64(chatID)
// 						// Удаляем клиента из комнаты
// 						if clients, ok := c.hub.rooms[chatIDInt]; ok {
// 							delete(clients, c)
// 							// Если комната пуста — удаляем её
// 							if len(clients) == 0 {
// 								delete(c.hub.rooms, chatIDInt)
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// func (c *Client) writePump() {
// 	defer func() {
// 		c.hub.unregister <- c
// 		c.conn.Close()
// 	}()
// 	for message := range c.send {
// 		err := c.conn.WriteMessage(websocket.TextMessage, message)
// 		if err != nil {
// 			break
// 		}
// 	}
// }

// func (m *MessagePayload) ToJSON() []byte {
// 	data, _ := json.Marshal(m)
// 	return data
// }

package ws

import (
	"chat-bd/models"
	"chat-bd/services"
	"chat-bd/utils"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients    map[*Client]bool
	rooms      map[int64]map[*Client]bool
	broadcast  chan MessagePayload
	register   chan *Client
	unregister chan *Client
	service    *services.ChatService
	mu         sync.RWMutex // Мьютекс для потокобезопасности
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	mu     sync.Mutex // Мьютекс для клиента
}

type MessagePayload struct {
	Event string         `json:"event"`
	Data  models.Message `json:"data"`
}

func NewHub(service *services.ChatService) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[int64]map[*Client]bool),
		broadcast:  make(chan MessagePayload, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		service:    service,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			log.Printf("Клиент зарегистрирован: UserID=%d, Всего клиентов: %d",
				client.userID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()

			// Удалить из clients
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)

				// Закрыть канал отправки (только если не закрыт)
				client.mu.Lock()
				select {
				case _, ok := <-client.send:
					if ok {
						close(client.send)
					}
				default:
					// Канал уже закрыт или пуст
				}
				client.mu.Unlock()

				log.Printf("Клиент удален: UserID=%d", client.userID)
			}

			// Удалить из всех rooms
			for chatID, clients := range h.rooms {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.rooms, chatID)
					}
				}
			}

			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clientsCopy := make([]*Client, 0, len(h.clients))
			for client := range h.clients {
				clientsCopy = append(clientsCopy, client)
			}
			h.mu.RUnlock()

			jsonData := message.ToJSON()
			for _, client := range clientsCopy {
				if client.userID == message.Data.SenderID {
					continue
				}

				client.mu.Lock()
				select {
				case client.send <- jsonData:
					// Успешно отправлено
				default:
					// Не удалось отправить - закрываем соединение
					go func(c *Client) {
						h.unregister <- c
						c.conn.Close()
					}(client)
				}
				client.mu.Unlock()
			}
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из query параметров
	authHeader := r.URL.Query().Get("token")

	if authHeader == "" {
		http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
		return
	}

	// Валидируем токен
	userID, err := utils.GetUserIDFromToken("Bearer " + authHeader)
	if err != nil {
		log.Printf("Ошибка валидации токена: %v", err)
		http.Error(w, "Недействительный токен", http.StatusUnauthorized)
		return
	}

	// Обновляем WebSocket с таймаутами
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка WebSocket: %v", err)
		return
	}

	// Устанавливаем таймауты
	conn.SetReadLimit(512 * 1024) // 512KB
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Создаем клиента
	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	// Регистрируем клиента
	h.register <- client

	// Получаем чаты пользователя
	chatIDs, err := h.service.GetChatIDsForUser(userID)
	if err != nil {
		log.Printf("Ошибка получения чатов для пользователя %d: %v", userID, err)
		chatIDs = []int64{}
	}

	// Подписываем на чаты
	h.mu.Lock()
	for _, chatID := range chatIDs {
		if _, ok := h.rooms[chatID]; !ok {
			h.rooms[chatID] = make(map[*Client]bool)
		}
		h.rooms[chatID][client] = true
		log.Printf("Пользователь %d подписан на чат %d", userID, chatID)
	}
	h.mu.Unlock()

	// Запускаем горутины
	go client.writePump()
	go client.readPump()

	log.Printf("WebSocket соединение установлено для UserID=%d", userID)
}

func (h *Hub) Broadcast(payload MessagePayload) {
	select {
	case h.broadcast <- payload:
		// Успешно добавлено в очередь
	default:
		log.Println("Очередь broadcast переполнена")
	}
}

func (h *Hub) BroadcastToRoom(chatID int64, payload MessagePayload) {
	h.mu.RLock()
	clients, ok := h.rooms[chatID]
	if !ok {
		h.mu.RUnlock()
		return
	}

	// Создаем копию для безопасной итерации
	clientsCopy := make([]*Client, 0, len(clients))
	for client := range clients {
		clientsCopy = append(clientsCopy, client)
	}
	h.mu.RUnlock()

	jsonData := payload.ToJSON()
	for _, client := range clientsCopy {
		if client.userID == payload.Data.SenderID {
			continue
		}

		client.mu.Lock()
		select {
		case client.send <- jsonData:
			// Успешно отправлено
		default:
			// Не удалось отправить
			go func(c *Client) {
				h.unregister <- c
				c.conn.Close()
			}(client)
		}
		client.mu.Unlock()
	}
}

func (c *Client) readPump() {
	defer func() {
		// Восстанавливаемся после паники
		if r := recover(); r != nil {
			log.Printf("Паника в readPump для UserID=%d: %v", c.userID, r)
		}

		// Отправляем unregister только один раз
		select {
		case c.hub.unregister <- c:
			// Успешно отправлено
		default:
			// Канал переполнен или hub остановлен
		}

		c.conn.Close()
		log.Printf("readPump завершен для UserID=%d", c.userID)
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				log.Printf("Ошибка чтения WebSocket для UserID=%d: %v", c.userID, err)
			}
			break
		}

		// Обработка сообщения
		var req map[string]interface{}
		if err := json.Unmarshal(message, &req); err != nil {
			log.Printf("Ошибка парсинга JSON для UserID=%d: %v", c.userID, err)
			continue
		}

		if event, ok := req["event"].(string); ok {
			switch event {
			case "join":
				if chatID, ok := req["chat_id"].(float64); ok {
					chatIDInt := int64(chatID)
					c.hub.mu.Lock()
					if _, ok := c.hub.rooms[chatIDInt]; !ok {
						c.hub.rooms[chatIDInt] = make(map[*Client]bool)
					}
					c.hub.rooms[chatIDInt][c] = true
					c.hub.mu.Unlock()
					log.Printf("UserID=%d присоединился к чату %d", c.userID, chatIDInt)
				}

			case "leave":
				if chatID, ok := req["chat_id"].(float64); ok {
					chatIDInt := int64(chatID)
					c.hub.mu.Lock()
					if clients, ok := c.hub.rooms[chatIDInt]; ok {
						delete(clients, c)
						if len(clients) == 0 {
							delete(c.hub.rooms, chatIDInt)
						}
					}
					c.hub.mu.Unlock()
					log.Printf("UserID=%d покинул чат %d", c.userID, chatIDInt)
				}

			case "ping":
				// Отправляем pong
				c.mu.Lock()
				select {
				case c.send <- []byte(`{"event":"pong"}`):
				default:
					// Не удалось отправить
				}
				c.mu.Unlock()
			}
		}

		// Сбрасываем таймаут чтения
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		// Восстанавливаемся после паники
		if r := recover(); r != nil {
			log.Printf("Паника в writePump для UserID=%d: %v", c.userID, r)
		}

		ticker.Stop()

		// Закрываем соединение
		c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()

		log.Printf("writePump завершен для UserID=%d", c.userID)
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// Канал закрыт
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Ошибка записи WebSocket для UserID=%d: %v", c.userID, err)
				return
			}

		case <-ticker.C:
			// Отправляем ping
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ошибка отправки ping для UserID=%d: %v", c.userID, err)
				return
			}
		}
	}
}

func (m *MessagePayload) ToJSON() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("Ошибка маршалинга MessagePayload: %v", err)
		return []byte(`{"event":"error","data":{"message":"internal error"}}`)
	}
	return data
}

func (h *Hub) SendToRoom(chatID int64, data []byte, excludeUserID int64) {
	h.mu.RLock()
	clients, ok := h.rooms[chatID]
	if !ok {
		h.mu.RUnlock()
		return
	}

	clientsCopy := make([]*Client, 0, len(clients))
	for client := range clients {
		clientsCopy = append(clientsCopy, client)
	}
	h.mu.RUnlock()

	for _, client := range clientsCopy {
		if client.userID == excludeUserID {
			continue
		}
		client.mu.Lock()
		select {
		case client.send <- data:
		default:
			go func(c *Client) {
				h.unregister <- c
				c.conn.Close()
			}(client)
		}
		client.mu.Unlock()
	}
}
