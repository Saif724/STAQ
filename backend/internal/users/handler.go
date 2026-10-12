package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
	"github.com/Saif724/STAQ/backend/internal/users/dto"
)

type Handler struct {
	service *Service
}

type updateProfileRequest struct {
	FullName string `json:"full_name"`
	Timezone string `json:"timezone"`
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
			"failed to retrieve user",
		)
		return
	}

	response.JSON(w, http.StatusOK, dto.MeResponse{
		ID:            user.ID,
		FullName:      user.FullName,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Timezone:      user.Timezone,
	})
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), userID, req.FullName, req.Timezone)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.ErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		}
		if errors.Is(err, ErrInvalidProfile) {
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_PROFILE", err.Error())
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "PROFILE_UPDATE_FAILED", "failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, dto.MeResponse{
		ID:            user.ID,
		FullName:      user.FullName,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Timezone:      user.Timezone,
	})
}
