package models

import "time"

type Chat struct {
	ID      int64     `json:"id"`
	User1ID int64     `json:"user1_id"`
	User2ID int64     `json:"user2_id"`
	Created time.Time `json:"created_at"`
}

type ChatPreview struct {
	ID            int64  `json:"id"`
	WithUserName  string `json:"with_user_name"`
	LastMessage   string `json:"last_message"`
	LastMessageAt int64  `json:"last_message_at"`
}

type Message struct {
	ID        int64  `json:"id"`
	ChatID    int64  `json:"chat_id"`
	SenderID  int64  `json:"sender_id"`
	Content   string `json:"content"`
	Read      bool   `json:"read"`
	CreatedAt int64  `json:"created_at"`
}
