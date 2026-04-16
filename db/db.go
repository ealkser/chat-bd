// db/db.go
package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./db.db")
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	err = createUsersTable()
	if err != nil {
		log.Fatal("Не удалось создать таблицу пользователей:", err)
	}

	err = createRefreshTokensTable()
	if err != nil {
		log.Fatal("Не удалось создать таблицу refresh_tokens:", err)
	}
}

func createUsersTable() error {
	sql := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := DB.Exec(sql)
	return err
}

func createRefreshTokensTable() error {
	sql := `CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	)`
	_, err := DB.Exec(sql)
	return err
}

// createTestUser — если хотите, можно оставить или удалить
func createTestUser() {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Println("Ошибка проверки пользователей:", err)
		return
	}

	if count == 0 {
		email := "test@example.com"
		name := "Test User"
		password := "password123"

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Ошибка хеширования пароля:", err)
			return
		}

		_, err = DB.Exec(
			"INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)",
			name, email, string(hashedPassword),
		)
		if err != nil {
			log.Println("Ошибка вставки тестового пользователя:", err)
		} else {
			log.Println("✅ Тестовый пользователь добавлен: test@example.com / password123")
		}
	}
}
