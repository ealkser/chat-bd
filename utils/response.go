// utils/response.go
package utils

import (
	"encoding/json"
	"net/http"
)

// Response стандартный формат ответа API
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondJSON отправляет JSON-ответ
func RespondJSON(w http.ResponseWriter, response Response, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
