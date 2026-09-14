package actions

import (
	"encoding/json"
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

type createActionRequest struct {
	TaskID            string          `json:"task_id"`
	ActionType        string          `json:"action_type"`
	ExecutionOrder    int             `json:"execution_order"`
	Configuration     json.RawMessage `json:"configuration"`
	ContinueOnFailure bool            `json:"continue_on_failure"`
}

type updateActionRequest struct {
	ActionType        string          `json:"action_type"`
	ExecutionOrder    int             `json:"execution_order"`
	Configuration     json.RawMessage `json:"configuration"`
	ContinueOnFailure bool            `json:"continue_on_failure"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createActionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	action, err := h.service.Create(
		r.Context(),
		userID,
		CreateActionInput{
			TaskID:            req.TaskID,
			ActionType:        req.ActionType,
			ExecutionOrder:    req.ExecutionOrder,
			Configuration:     req.Configuration,
			ContinueOnFailure: req.ContinueOnFailure,
		},
	)

	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, action)
}

func (h *Handler) ListByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("taskID")
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

	actions, err := h.service.ListByTaskID(r.Context(), taskID, userID)

	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, actions)
}

func (h *Handler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	actionID := r.PathValue("id")
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

	action, err := h.service.GetByID(r.Context(), actionID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, action)
}

func (h *Handler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	actionID := r.PathValue("id")
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

	var req updateActionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	trigger, err := h.service.Update(
		r.Context(),
		userID,
		UpdateActionInput{
			ID:                actionID,
			ActionType:        req.ActionType,
			ExecutionOrder:    req.ExecutionOrder,
			Configuration:     req.Configuration,
			ContinueOnFailure: req.ContinueOnFailure,
		},
	)

	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, trigger)
}

func (h *Handler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	actionID := r.PathValue("id")
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

	if err := h.service.Delete(
		r.Context(),
		actionID,
		userID,
	); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrActionNotFound):
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"ACTION_NOT_FOUND",
			"action not found",
		)

	case errors.Is(err, ErrInvalidActionType):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_ACTION_TYPE",
			"invalid action type",
		)

	case errors.Is(err, ErrInvalidExecutionOrder):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_EXECUTION_ORDER",
			"execution order must be greater than zero",
		)

	case errors.Is(err, ErrInvalidConfiguration):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_CONFIGURATION",
			"invalid action configuration",
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
