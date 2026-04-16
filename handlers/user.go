// handlers/user.go
package handlers

import (
	"chat-bd/models"
	"chat-bd/utils"
	"net/http"
)

func MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Метод не поддерживается"}, http.StatusMethodNotAllowed)
		return
	}

	tokenString := r.Header.Get("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	} else {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Требуется токен Bearer"}, http.StatusUnauthorized)
		return
	}

	claims, valid := utils.ValidateJWT(tokenString)
	if !valid {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Невалидный или просроченный токен"}, http.StatusUnauthorized)
		return
	}

	userID := int((*claims)["sub"].(float64))

	var user models.User
	err := user.FindByID(userID)
	if err != nil {
		utils.RespondJSON(w, utils.Response{Success: false, Message: "Пользователь не найден"}, http.StatusNotFound)
		return
	}

	utils.RespondJSON(w, utils.Response{
		Success: true,
		Message: "Данные пользователя",
		Data:    user,
	}, http.StatusOK)
}
