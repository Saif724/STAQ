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

		case errors.Is(err, ErrEmailNotVerified):
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

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	result, err := h.service.RefreshAccessToken(
		r.Context(),
		req.RefreshToken,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken):
			response.ErrorJSON(
				w,
				http.StatusUnauthorized,
				"INVALID_REFRESH_TOKEN",
				"invalid refresh token",
			)
		default:
			response.ErrorJSON(
				w,
				http.StatusInternalServerError,
				"REFRESH_FAILED",
				"failed to refresh access token",
			)
		}
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken):
			response.ErrorJSON(
				w,
				http.StatusUnauthorized,
				"INVALID_REFRESH_TOKEN",
				"invalid refresh token",
			)
		default:
			response.ErrorJSON(
				w,
				http.StatusInternalServerError,
				"LOGOUT_FAILED",
				"failed to logout",
			)
		}

		return
	}

	response.JSON(
		w,
		http.StatusOK,
		"logged out successfully",
	)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	err := h.service.VerifyEmail(
		r.Context(),
		req.Email,
		req.Code,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyVerified):
			response.ErrorJSON(
				w,
				http.StatusConflict,
				"EMAIL_ALREADY_VERIFIED",
				"email is already verified",
			)

		case errors.Is(err, ErrVerificationExpired):
			response.ErrorJSON(
				w,
				http.StatusUnauthorized,
				"VERIFICATION_EXPIRED",
				"verification code has expired",
			)

		case errors.Is(err, ErrInvalidVerificationCode):
			response.ErrorJSON(
				w,
				http.StatusUnauthorized,
				"INVALID_VERIFICATION_CODE",
				"invalid verification code",
			)

		default:
			response.ErrorJSON(
				w,
				http.StatusInternalServerError,
				"VERIFICATION_FAILED",
				"failed to verify email",
			)
		}

		return
	}

	response.JSON(
		w,
		http.StatusOK,
		dto.VerifyEmailResponse{
			Message: "email verified successfully",
		},
	)
}

func (h *Handler) ResendVerification(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req dto.ResendVerificationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	if err := h.service.ResendVerificationEmail(
		r.Context(),
		req.Email,
	); err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyVerified):
			response.ErrorJSON(
				w,
				http.StatusConflict,
				"EMAIL_ALREADY_VERIFIED",
				"email is already verified",
			)

		case errors.Is(err, ErrInvalidVerificationCode):
			response.ErrorJSON(
				w,
				http.StatusUnauthorized,
				"INVALID_VERIFICATION",
				"unable to resend verification email",
			)

		default:
			response.ErrorJSON(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"failed to resend verification email",
			)
		}

		return
	}

	response.JSON(
		w,
		http.StatusOK,
		dto.ResendVerificationResponse{
			Message: "verification email sent",
		},
	)
}
