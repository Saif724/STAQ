package tasksetups

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "user authentication is required")
		return
	}

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	result, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidSetup) {
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_TASK_SETUP", err.Error())
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "TASK_SETUP_FAILED", "failed to create task setup")
		return
	}

	response.JSON(w, http.StatusCreated, result)
}
