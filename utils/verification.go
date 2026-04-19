package utils

import (
	"chat-bd/db"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

// GenerateVerificationCode создаёт 6-значный код и сохраняет в БД
func GenerateVerificationCode(email string) (string, error) {
	code := fmt.Sprintf("%06d", rand.Intn(999999))
	expiresAt := time.Now().Add(10 * time.Minute) // 10 минут

	query := `
		INSERT INTO verification_codes (email, code, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE
		SET code = $2, expires_at = $3`

	_, err := db.DB.Exec(query, email, code, expiresAt)
	if err != nil {
		return "", fmt.Errorf("не удалось сохранить код: %w", err)
	}

	return code, nil
}

// ValidateVerificationCode проверяет, совпадает ли код и не просрочен ли
func ValidateVerificationCode(email, code string) (bool, error) {
	var storedCode string
	var expiresAt time.Time

	query := `SELECT code, expires_at FROM verification_codes WHERE email = $1`
	err := db.DB.QueryRow(query, email).Scan(&storedCode, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if time.Now().After(expiresAt) {
		return false, nil
	}

	return storedCode == code, nil
}

// ClearVerificationCode удаляет код после успешной верификации
func ClearVerificationCode(email string) error {
	query := `DELETE FROM verification_codes WHERE email = $1`
	_, err := db.DB.Exec(query, email)
	return err
}
