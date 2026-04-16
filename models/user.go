// models/user.go
package models

import (
	"chat-bd/db"
	_ "database/sql"

	"golang.org/x/crypto/bcrypt"
)

// User представляет пользователя
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Password не сохраняется в БД напрямую
	Password string `json:"-"`
}

// HashPassword хеширует пароль
func (u *User) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

// CheckPassword проверяет пароль
func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

// Create сохраняет пользователя в БД
func (u *User) Create() error {
	err := u.HashPassword()
	if err != nil {
		return err
	}

	// Используем Exec + LastInsertId()
	result, err := db.DB.Exec(
		"INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)",
		u.Name, u.Email, u.Password,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	u.ID = int(id) // ✅ ВАЖНО: сохраняем ID
	return nil
}

// FindByEmail ищет пользователя по email
func (u *User) FindByEmail(email string) error {
	var passwordHash string
	err := db.DB.QueryRow(
		"SELECT id, name, email, password_hash FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &passwordHash)
	if err != nil {
		return err
	}
	u.Password = passwordHash // Хеш для проверки
	return nil
}

// FindByID находит пользователя по ID
func (u *User) FindByID(id int) error {
	return db.DB.QueryRow(
		"SELECT id, name, email FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Name, &u.Email)
}
