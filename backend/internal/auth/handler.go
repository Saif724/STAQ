package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/auth/dto"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
	"github.com/Saif724/STAQ/backend/internal/users"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	result, err := h.service.Register(r.Context(), req)

	if err != nil {
		if errors.Is(err, users.ErrEmailExists) {
			response.ErrorJSON(
				w,
				http.StatusConflict,
				"EMAIL_EXISTS",
				"an account with this email already exists",
			)
			return
		}
		response.ErrorJSON(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)

		return
	}

	result, err := h.service.Login(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			response.ErrorJSON(
				w, http.StatusUnauthorized,
				"INVALID_CREDENTIALS",
				"invalid email or password",
			)

		case errors.Is(err, ErrAccountInactive):
			response.ErrorJSON(
				w,
				http.StatusForbidden,
				"ACCOUNT_INACTIVE",
				"account is inactive",
			)

		case errors.Is(err, ErrEamilNotVerified):
			response.ErrorJSON(
				w,
				http.StatusForbidden,
				"EMAIL_NOT_VERIFIED",
				"email is not verified",
			)

		default:
			response.ErrorJSON(
				w,
				http.StatusBadRequest,
				"LOGIN_FAILED",
				"login failed",
			)
		}
		return
	}

	response.JSON(w, http.StatusOK, result)
}
