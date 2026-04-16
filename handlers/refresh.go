// handlers/refresh.go
package handlers

import (
	"encoding/json"
	"net/http"

	"chat-bd/models"
	"chat-bd/utils"
)

// RefreshHandler выдаёт новый access token по refresh token
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Метод не поддерживается"}, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Некорректные данные"}, http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Требуется refresh_token"}, http.StatusBadRequest)
		return
	}

	userID, valid := utils.ValidateRefreshToken(req.RefreshToken)
	if !valid {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Невалидный или просроченный refresh token"}, http.StatusUnauthorized)
		return
	}

	// Генерация нового access token
	var user models.User
	if err := user.FindByID(userID); err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Пользователь не найден"}, http.StatusNotFound)
		return
	}
	newAccessToken, err := utils.GenerateJWT(userID, user.Email)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Ошибка генерации access token"}, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, utils.Response{
		Success: true,
		Message: "Token refreshed",
		Data: map[string]string{
			"access_token": newAccessToken,
		},
	}, http.StatusOK)
}
