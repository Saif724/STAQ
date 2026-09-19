package email

import (
	"net/http"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) BeginGmailAuthorization(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"Authentication is required",
		)

		return
	}

	authURL, err := h.service.BeginGmailAuthorization(r.Context(), userID)
	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"GMAIL_AUTHORIZATION_FAILED",
			"Failed to start Gmail authorization",
		)

		return
	}

	http.Redirect(
		w,
		r,
		authURL,
		http.StatusTemporaryRedirect,
	)
}

func (h *Handler) CompleteGmailAuthorization(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query()

	state := strings.TrimSpace(query.Get("state"))
	code := strings.TrimSpace(query.Get("code"))
	oauthError := strings.TrimSpace(query.Get("error"))

	if oauthError != "" {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"GMAIL_AUTHORIZATION_DENIED",
			"Google authorization was denied",
		)

		return
	}

	if state == "" || code == "" {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_OAUTH_CALLBACK",
			"Missing OAuth state or authorization code",
		)
	}

	userID, connection, err := h.service.CompleteGmailAuthorization(r.Context(), state, code)

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"GMAIL_CONNECTION_FAILED",
			"Failed to connect Gmail account",
		)
		return
	}

	if connection == nil {
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"GMAIL_CONNECTION_FAILED",
			"Gmail connection was not created",
		)
	}

	response.JSON(
		w,
		http.StatusOK,
		map[string]any{
			"user_id":       userID,
			"connection":    connection,
			"provider":      connection.Provider,
			"account_email": connection.AccountEmail,
		},
	)
}
