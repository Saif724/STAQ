package connections

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
)

type Handler struct {
	service *Service
}

type ConnectionResponse struct {
	ID           string     `json:"id"`
	Provider     string     `json:"provider"`
	AccountEmail string     `json:"account_email"`
	Scopes       []string   `json:"scopes"`
	TokenExpires *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "authenticated user not found")
		return
	}

	connections, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "CONNECTION_LIST_FAILED", "failed to list connections")
		return
	}

	items := make([]ConnectionResponse, 0, len(connections))
	for _, connection := range connections {
		items = append(items, toConnectionResponse(connection))
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "authenticated user not found")
		return
	}

	connectionID := strings.TrimSpace(r.PathValue("id"))
	if connectionID == "" {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_CONNECTION_ID", "connection id is required")
		return
	}

	if err := h.service.Delete(r.Context(), connectionID, userID); err != nil {
		if errors.Is(err, ErrConnectionNotFound) {
			response.ErrorJSON(w, http.StatusNotFound, "CONNECTION_NOT_FOUND", "connection not found")
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "CONNECTION_DELETE_FAILED", "failed to disconnect account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toConnectionResponse(connection *Connection) ConnectionResponse {
	return ConnectionResponse{
		ID:           connection.ID,
		Provider:     connection.Provider,
		AccountEmail: connection.AccountEmail,
		Scopes:       connection.Scopes,
		TokenExpires: connection.TokenExpiresAt,
		CreatedAt:    connection.CreatedAt,
		UpdatedAt:    connection.UpdatedAt,
	}
}
