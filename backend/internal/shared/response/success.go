package response

import (
	"encoding/json"
	"net/http"
)

type DataResponse struct {
	Data any `json:"data"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := DataResponse{
		Data: data,
	}

	_ = json.NewEncoder(w).Encode(response)
}

func Message(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := MessageResponse{
		Message: message,
	}

	_ = json.NewEncoder(w).Encode(response)
}