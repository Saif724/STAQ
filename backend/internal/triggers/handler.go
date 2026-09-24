package triggers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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

type createTriggerRequest struct {
	TaskID         string    `json:"task_id"`
	TriggerType    string    `json:"trigger_type"`
	CronExpression *string   `json:"cron_expression"`
	Timezone       string    `json:"timezone"`
	StartAt        time.Time `json:"start_at"`
}

type updateTriggerRequest struct {
	TriggerType    string    `json:"trigger_type"`
	CronExpression *string   `json:"cron_expression"`
	Timezone       string    `json:"timezone"`
	StartAt        time.Time `json:"start_at"`
	IsActive       bool      `json:"is_active"`
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

	var req createTriggerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	trigger, err := h.service.Create(
		r.Context(),
		userID,
		CreateTriggerInput{
			TaskID:         req.TaskID,
			TriggerType:    req.TriggerType,
			CronExpression: req.CronExpression,
			Timezone:       req.Timezone,
			StartAt:        req.StartAt,
		},
	)

	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, trigger)
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

	triggers, err := h.service.ListByTaskID(r.Context(), taskID, userID)

	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, triggers)
}

func (h *Handler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	triggerID := r.PathValue("id")
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

	trigger, err := h.service.GetByID(r.Context(), triggerID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, trigger)
}

func (h *Handler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	triggerID := r.PathValue("id")
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

	var req updateTriggerRequest

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
		UpdateTriggerInput{
			ID:             triggerID,
			TriggerType:    req.TriggerType,
			CronExpression: req.CronExpression,
			Timezone:       req.Timezone,
			StartAt:        req.StartAt,
			IsActive:       req.IsActive,
		},
		userID,
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
	triggerID := r.PathValue("id")
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
		triggerID,
		userID,
	); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrTriggerNotFound):
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"TRIGGER_NOT_FOUND",
			"trigger not found",
		)

	case errors.Is(err, ErrInvalidTriggerType):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_TRIGGER_TYPE",
			"invalid trigger type",
		)

	case errors.Is(err, ErrInvalidTimezone):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_TIMEZONE",
			"invalid timezone",
		)

	case errors.Is(err, ErrInvalidStartAt):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_START_AT",
			"start_at is required",
		)

	case errors.Is(err, ErrStartAtInPast):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"START_AT_IN_PAST",
			"start_at must be in the future for ONCE triggers",
		)

	case errors.Is(err, ErrCronRequired):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"CRON_REQUIRED",
			"cron expression is required for CRON triggers",
		)

	case errors.Is(err, ErrCronInvalid):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_CRON",
			"invalid cron expression",
		)

	case strings.Contains(err.Error(), "cron expression is only allowed"):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"CRON_NOT_ALLOWED",
			"cron expression is only allowed for CRON triggers",
		)

	case strings.Contains(err.Error(), "task not found"):
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"TASK_NOT_FOUND",
			"task not found",
		)

	case strings.Contains(err.Error(), "task id is required"):
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_TASK",
			"task id is required",
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
