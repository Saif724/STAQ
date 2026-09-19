package email

import (
	"net/http"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/connections"
	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
)

type connectionResponse struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	Provider          string    `json:"provider"`
	ProviderAccountID string    `json:"provider_account_id"`
	AccountEmail      string    `json:"account_email"`
	Scopes            []string  `json:"scopes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

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
			"connection":    toConnectionResponst(connection),
			"provider":      connection.Provider,
			"account_email": connection.AccountEmail,
		},
	)
}

func toConnectionResponst(
	connection *connections.Connection,
) connectionResponse {
	return connectionResponse{
		ID:                connection.ID,
		UserID:            connection.UserID,
		Provider:          connection.Provider,
		ProviderAccountID: connection.ProviderAccountID,
		AccountEmail:      connection.AccountEmail,
		Scopes:            connection.Scopes,
		CreatedAt:         connection.CreatedAt,
		UpdatedAt:         connection.UpdatedAt,
	}
}
