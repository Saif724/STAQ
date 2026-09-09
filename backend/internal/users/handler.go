package users

import (
	"errors"
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
	"github.com/Saif724/STAQ/backend/internal/users/dto"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())

	if !ok || userID == "" {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"unauthorized",
		)
		return
	}

	user, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.ErrorJSON(
				w,
				http.StatusNotFound,
				"USER_NOT_FOUND",
				"user not found",
			)
			return
		}

		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"USER_LOOKUP_FAILED",
			"failed to retrive user",
		)
		return
	}

	response.JSON(w, http.StatusOK, dto.MeResponse{
		ID:            user.ID,
		FullName:      user.FullName,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	})
}
