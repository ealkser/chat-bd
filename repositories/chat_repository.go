package repositories

import (
	"chat-bd/models"
	"database/sql"
	"fmt"
)

type ChatResponse struct {
	ID            int64  `json:"id"`
	WithUserName  string `json:"with_user_name"`
	LastMessage   string `json:"last_message"`
	LastMessageAt int64  `json:"last_message_at"`
}

type MessageResponse struct {
	ID        int64  `json:"id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	Read      bool   `json:"read"`
	CreatedAt int64  `json:"created_at"`
}

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// GetChatsForUser возвращает список чатов с последним сообщением
func (r *ChatRepository) GetChatsForUser(userID int64) ([]models.ChatPreview, error) {
	query := `
		SELECT 
			c.id,
			u.name AS with_user_name,
			(SELECT m.content 
			 FROM messages m 
			 WHERE m.chat_id = c.id 
			 ORDER BY m.created_at DESC 
			 LIMIT 1),
			COALESCE(
				(SELECT strftime('%s', m.created_at)
				 FROM messages m 
				 WHERE m.chat_id = c.id 
				 ORDER BY m.created_at DESC 
				 LIMIT 1),
				strftime('%s', c.created_at)
			) AS last_message_at
		FROM chats c
		JOIN users u ON (u.id = c.user1_id OR u.id = c.user2_id) AND u.id != $1
		WHERE $1 IN (c.user1_id, c.user2_id)
		ORDER BY last_message_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var chats []models.ChatPreview
	for rows.Next() {
		var chat models.ChatPreview
		err := rows.Scan(
			&chat.ID,
			&chat.WithUserName,
			&chat.LastMessage,
			&chat.LastMessageAt, // ← сканируем как int64
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

// GetMessages возвращает все сообщения из чата
func (r *ChatRepository) GetMessages(chatID int64) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, content, read, strftime('%s', created_at)
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		err := rows.Scan(&m.ID, &m.SenderID, &m.Content, &m.Read, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования сообщения: %w", err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}

// CreateMessage отправляет новое сообщение
func (r *ChatRepository) CreateMessage(chatID, senderID int64, content string) (int64, error) {
	var messageID int64
	query := `
		INSERT INTO messages (chat_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err := r.db.QueryRow(query, chatID, senderID, content).Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("ошибка отправки сообщения: %w", err)
	}
	return messageID, nil
}

// GetOrCreateChat находит или создаёт чат между двумя пользователями
func (r *ChatRepository) GetOrCreateChat(user1ID, user2ID int64) (int64, error) {
	if user1ID == user2ID {
		return 0, fmt.Errorf("нельзя создать чат с самим собой")
	}

	// Упорядочиваем ID, чтобы избежать дублей
	a, b := user1ID, user2ID
	if a > b {
		a, b = b, a
	}

	var chatID int64
	err := r.db.QueryRow(`
		INSERT INTO chats (user1_id, user2_id)
		VALUES ($1, $2)
		ON CONFLICT (user1_id, user2_id) DO NOTHING
		RETURNING id
	`, a, b).Scan(&chatID)

	if err != nil {
		return 0, fmt.Errorf("ошибка создания чата: %w", err)
	}

	// Если чат уже существует — получаем его ID
	if err == sql.ErrNoRows {
		err = r.db.QueryRow("SELECT id FROM chats WHERE user1_id = $1 AND user2_id = $2", a, b).Scan(&chatID)
		if err != nil {
			return 0, fmt.Errorf("ошибка получения существующего чата: %w", err)
		}
	}

	return chatID, nil
}

// UserExists проверяет, существует ли пользователь
func (r *ChatRepository) UserExists(email string) (int64, bool, error) {
	var id int64
	err := r.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("ошибка проверки пользователя: %w", err)
	}
	return id, true, nil
}

// IsUserInChat проверяет, состоит ли пользователь в чате
func (r *ChatRepository) IsUserInChat(chatID, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM chats
			WHERE id = $1 AND $2 IN (user1_id, user2_id)
		)
	`, chatID, userID).Scan(&exists)
	return exists, err
}

func (r *ChatRepository) GetChatIDsForUser(userID int64) ([]int64, error) {
	query := `
		SELECT id FROM chats
		WHERE user1_id = ? OR user2_id = ?
		ORDER BY id
	`

	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("ошибка сканирования chat_id: %w", err)
		}
		chatIDs = append(chatIDs, id)
	}

	return chatIDs, nil
}
