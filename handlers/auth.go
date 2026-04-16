// handlers/auth.go
package handlers

import (
	"chat-bd/models"
	"chat-bd/utils"
	"encoding/json"
	"net/http"
	"time"
)

// Response представляет стандартный ответ API
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RegisterHandler обрабатывает регистрацию нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Метод не поддерживается"}, http.StatusMethodNotAllowed)
		return
	}

	// Используем временную структуру, чтобы пароль НЕ игнорировался
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Некорректные данные"}, http.StatusBadRequest)
		return
	}

	if creds.Email == "" || creds.Password == "" || creds.Name == "" {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Email, Name и пароль обязательны"}, http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь
	var existing models.User
	err := existing.FindByEmail(creds.Email)
	if err == nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Пользователь с таким email уже существует"}, http.StatusConflict)
		return
	}

	// Создаём нового пользователя
	user := models.User{
		Name:     creds.Name,
		Email:    creds.Email,
		Password: creds.Password,
	}
	if err := user.Create(); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка регистрации"}, http.StatusInternalServerError)
		return
	}

	// Генерация access token
	accessToken, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка генерации access token"}, http.StatusInternalServerError)
		return
	}

	// Генерация refresh token
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка генерации refresh token"}, http.StatusInternalServerError)
		return
	}

	// Сохраняем refresh token в БД
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = utils.StoreRefreshToken(user.ID, refreshToken, expiresAt)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка сохранения refresh token"}, http.StatusInternalServerError)
		return
	}

	// Отправляем оба токена
	utils.RespondJSON(w, utils.Response{
		Success: true,
		Message: "Регистрация успешна",
		Data: map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	}, http.StatusCreated)
}

// LoginHandler обрабатывает аутентификацию пользователя
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Метод не поддерживается"}, http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Некорректные данные"}, http.StatusBadRequest)
		return
	}

	if creds.Email == "" || creds.Password == "" {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Email и пароль обязательны"}, http.StatusBadRequest)
		return
	}

	var user models.User
	if err := user.FindByEmail(creds.Email); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неверные учетные данные"}, http.StatusUnauthorized)
		return
	}

	if !user.CheckPassword(creds.Password) {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неверные учетные данные"}, http.StatusUnauthorized)
		return
	}

	// Генерация access token
	accessToken, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка генерации access token"}, http.StatusInternalServerError)
		return
	}

	// Генерация refresh token
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка генерации refresh token"}, http.StatusInternalServerError)
		return
	}

	// Сохраняем refresh token в БД
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = utils.StoreRefreshToken(user.ID, refreshToken, expiresAt)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка сохранения refresh token"}, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, utils.Response{
		Success: true,
		Message: "Аутентификация успешна",
		Data: map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	}, http.StatusOK)
}

// LogoutHandler удаляет refresh token из БД
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Метод не поддерживается"}, http.StatusMethodNotAllowed)
		return
	}

	// Клиент должен прислать refresh_token
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.RefreshToken == "" {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Требуется refresh_token"}, http.StatusBadRequest)
		return
	}

	// Удаляем из БД (отзываем)
	utils.RevokeRefreshToken(req.RefreshToken)

	utils.RespondJSON(w, utils.Response{
		Success: true,
		Message: "Выход успешен",
	}, http.StatusOK)
}
