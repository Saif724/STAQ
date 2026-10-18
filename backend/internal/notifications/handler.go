package notifications

import (
	"errors"
	"net/http"
	"strconv"
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

func currentUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required")
		return "", false
	}
	return userID, true
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit")) // invalid -> 0 -> service default

	result, err := h.service.List(r.Context(), userID, limit)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "NOTIFICATIONS_FAILED", "Failed to load notifications")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	err := h.service.MarkRead(r.Context(), r.PathValue("id"), userID)
	if errors.Is(err, ErrNotFound) {
		response.ErrorJSON(w, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Notification not found")
		return
	}
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "NOTIFICATIONS_FAILED", "Failed to update notification")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"read": true})
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	if err := h.service.MarkAllRead(r.Context(), userID); err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "NOTIFICATIONS_FAILED", "Failed to update notifications")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"updated": true})
}
