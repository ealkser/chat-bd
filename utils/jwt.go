// utils/jwt.go
package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(userID int, email string) (string, error) {
	claims := &jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(30 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.MapClaims, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return &claims, true
	}

	return nil, false
}

const BearerSchema = "Bearer "

// GetUserIDFromToken извлекает user_id из JWT токена в заголовке Authorization
func GetUserIDFromToken(authHeader string) (int64, error) {
	if authHeader == "" {
		return 0, fmt.Errorf("отсутствует токен авторизации")
	}

	if !strings.HasPrefix(authHeader, BearerSchema) {
		return 0, fmt.Errorf("некорректный формат токена")
	}

	tokenString := strings.TrimPrefix(authHeader, BearerSchema)

	// Используем общий ValidateJWT
	claims, valid := ValidateJWT(tokenString)
	if !valid {
		return 0, fmt.Errorf("неверный или просроченный токен")
	}

	// Извлекаем "sub" (subject) — это наш userID
	sub, exists := (*claims)["sub"]
	if !exists {
		return 0, fmt.Errorf("отсутствует идентификатор пользователя в токене")
	}

	// JWT числа приходят как float64
	userIDFloat, ok := sub.(float64)
	if !ok {
		return 0, fmt.Errorf("неверный формат идентификатора пользователя")
	}

	return int64(userIDFloat), nil
}
