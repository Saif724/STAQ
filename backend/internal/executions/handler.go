package executions

import (
	"errors"
	"net/http"

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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	executionID := r.PathValue("id")

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authenticated user not found",
		)
		return
	}

	execution, err := h.service.GetByIDAndUser(r.Context(), executionID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		execution,
	)
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	executionID := r.PathValue("id")

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authenticated user not found",
		)
		return
	}

	logs, err := h.service.ListLogsByUser(r.Context(), executionID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		logs,
	)
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrExecutionNotFound):
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"EXECUTION_NOT_FOUND",
			"execution not found",
		)

	default:
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		)
	}
}
