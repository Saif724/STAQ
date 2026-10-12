package dashboard

import (
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "authenticated user not found")
		return
	}

	summary, err := h.service.GetForUser(r.Context(), userID)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "DASHBOARD_SUMMARY_FAILED", "failed to load dashboard summary")
		return
	}

	response.JSON(w, http.StatusOK, summary)
}
