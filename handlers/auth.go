// handlers/auth.go
package handlers

import (
	"chat-bd/models"
	"chat-bd/services"
	"chat-bd/utils"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var emailService *services.EmailService

func init() {
	emailService = services.NewEmailService(
		"nekstcat@gmail.com",
		"rvgp rubq xupm qjwq", // app password
		"smtp.gmail.com",
		"587",
	)
}

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

	// Генерируем и сохраняем код в БД
	code, err := utils.GenerateVerificationCode(creds.Email) // ← ваш db из глобального подключения
	if err != nil {
		log.Printf("Ошибка генерации кода: %v", err)
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Внутренняя ошибка сервера"}, http.StatusInternalServerError)
		return
	}

	// Отправляем код на email
	err = emailService.SendVerificationCode(creds.Email, creds.Name, code)
	if err != nil {
		log.Printf("Ошибка отправки email: %v", err)
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Не удалось отправить код"}, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, utils.Response{
		Success: true,
		Message: "Код подтверждения отправлен",
		Data:    map[string]string{"email": creds.Email},
	}, http.StatusOK)
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
			"name":          user.Name,
			"email":         user.Email,
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

func VerifyCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Метод не поддерживается"}, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Некорректные данные"}, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Code == "" {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Email и код обязательны"}, http.StatusBadRequest)
		return
	}

	// Проверяем код в БД
	valid, err := utils.ValidateVerificationCode(req.Email, req.Code)
	if err != nil {
		log.Printf("Ошибка проверки кода: %v", err)
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Внутренняя ошибка"}, http.StatusInternalServerError)
		return
	}
	if !valid {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Неверный или просроченный код"}, http.StatusUnauthorized)
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
	if err := user.Create(); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка создания пользователя"}, http.StatusInternalServerError)
		return
	}

	// Удаляем код
	utils.ClearVerificationCode(req.Email)

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
		Message: "Регистрация завершена",
		Data: map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"name":          user.Name,
			"email":         user.Email,
		},
	}, http.StatusOK)
}
