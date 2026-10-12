package email

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/shared/response"
)

type Handler struct {
	service     *Service
	frontendURL string
}

func NewHandler(service *Service, frontendURL string) *Handler {
	return &Handler{
		service:     service,
		frontendURL: strings.TrimRight(frontendURL, "/"),
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

func (h *Handler) GmailAuthorizationURL(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required")
		return
	}

	authorizationURL, err := h.service.BeginGmailAuthorization(r.Context(), userID)
	if err != nil {
		response.ErrorJSON(w, http.StatusInternalServerError, "GMAIL_AUTHORIZATION_FAILED", "Failed to start Gmail authorization")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"authorization_url": authorizationURL,
	})
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
		h.redirectGmailResult(w, r, "error", "denied")
		return
	}

	if state == "" || code == "" {
		h.redirectGmailResult(w, r, "error", "invalid_callback")
		return
	}

	_, connection, err := h.service.CompleteGmailAuthorization(r.Context(), state, code)

	if err != nil || connection == nil {
		h.redirectGmailResult(w, r, "error", "connection_failed")
		return
	}

	h.redirectGmailResult(w, r, "connected", "")
}

func (h *Handler) redirectGmailResult(
	w http.ResponseWriter,
	r *http.Request,
	result string,
	reason string,
) {
	frontend, err := url.Parse(h.frontendURL)
	if err != nil || frontend.Host == "" || (frontend.Scheme != "https" && frontend.Scheme != "http") {
		response.ErrorJSON(w, http.StatusInternalServerError, "GMAIL_CALLBACK_FAILED", "Gmail authorization could not be completed")
		return
	}

	frontend.Path = "/dashboard/connections"
	query := frontend.Query()
	query.Set("gmail", result)
	if reason != "" {
		query.Set("reason", reason)
	}
	frontend.RawQuery = query.Encode()
	http.Redirect(w, r, frontend.String(), http.StatusSeeOther)
}
