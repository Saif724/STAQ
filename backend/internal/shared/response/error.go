package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code string `json:"code"`
	Message string `json:"message"`
}

func ErrorJSON(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Error: Error{
			Code: code,
			Message: message,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}
