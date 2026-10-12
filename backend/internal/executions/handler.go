package executions

import (
	"errors"
	"net/http"
	"strconv"
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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

	page, err := parsePositiveInt64(r.URL.Query().Get("page"), 1)
	if err != nil {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_PAGINATION", "page must be a positive integer")
		return
	}
	pageSize, err := parsePositiveInt64(r.URL.Query().Get("page_size"), DefaultExecutionPageSize)
	if err != nil || pageSize > MaxExecutionPageSize {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_PAGINATION", "page_size must be between 1 and 100")
		return
	}

	pageResult, err := h.service.ListByUser(
		r.Context(),
		userID,
		r.URL.Query().Get("task_id"),
		strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status"))),
		page,
		pageSize,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidExecutionFilter) {
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_EXECUTION_STATUS", err.Error())
			return
		}
		if errors.Is(err, ErrInvalidPagination) {
			response.ErrorJSON(w, http.StatusBadRequest, "INVALID_PAGINATION", err.Error())
			return
		}
		response.ErrorJSON(w, http.StatusInternalServerError, "EXECUTION_LIST_FAILED", "failed to list executions")
		return
	}

	response.JSON(w, http.StatusOK, pageResult)
}

func parsePositiveInt64(value string, fallback int64) (int64, error) {
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 1 {
		return 0, errors.New("value must be a positive integer")
	}

	return parsed, nil
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
