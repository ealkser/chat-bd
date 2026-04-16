// utils/refreshtoken.go
package utils

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"chat-bd/db"
)

const refreshTokenLength = 32

// GenerateRefreshToken создаёт случайный токен
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, refreshTokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// StoreRefreshToken сохраняет токен в БД
func StoreRefreshToken(userID int, token string, expiresAt time.Time) error {
	_, err := db.DB.Exec(
		"INSERT OR REPLACE INTO refresh_tokens (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	return err
}

// ValidateRefreshToken проверяет, существует ли токен и не просрочен ли
func ValidateRefreshToken(token string) (int, bool) {
	var userID int
	var expiresAt time.Time

	err := db.DB.QueryRow(
		"SELECT user_id, expires_at FROM refresh_tokens WHERE token = ?",
		token,
	).Scan(&userID, &expiresAt)

	if err == sql.ErrNoRows || err != nil {
		return 0, false
	}

	if time.Now().After(expiresAt) {
		return 0, false
	}

	return userID, true
}

// RevokeRefreshToken удаляет токен (при logout)
func RevokeRefreshToken(token string) error {
	_, err := db.DB.Exec("DELETE FROM refresh_tokens WHERE token = ?", token)
	return err
}