package tasks

import (
	"encoding/json"
	"errors"
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

type createTaskRequest struct {
	QueueID        string  `json:"queue_id"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	MaxRetries     int     `json:"max_retries"`
}

type updateTaskRequest struct {
	QueueID        string  `json:"queue_id"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	Status         string  `json:"status"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	MaxRetries     int     `json:"max_retries"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())

	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"user authentication is required",
		)
		return
	}

	var req createTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	task, err := h.service.Create(r.Context(), CreateTaskInput{
		UserID:         userID,
		QueueID:        req.QueueID,
		Name:           req.Name,
		Description:    req.Description,
		TimeoutSeconds: req.TimeoutSeconds,
		MaxRetries:     req.MaxRetries,
	})

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"TASK_CREATION_FAILED",
			err.Error(),
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		task,
	)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())

	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"user authentication is required",
		)
		return
	}

	tasks, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"TASK_LIST_FAILED",
			err.Error(),
		)
		return
	}

	response.JSON(
		w,
		http.StatusOK,
		tasks,
	)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())

	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"user authentication is required",
		)
		return
	}

	taskID := r.PathValue("id")

	task, err := h.service.GetByID(
		r.Context(),
		taskID,
		userID,
	)

	if errors.Is(err, ErrTaskNotFound) {
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"TASK_NOT_FOUND",
			"task not found",
		)
		return
	}

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"TASK_FETCH_FAILED",
			err.Error(),
		)

		return
	}

	response.JSON(
		w,
		http.StatusOK,
		task,
	)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())

	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"user authentication is required",
		)
		return
	}

	taskID := r.PathValue("id")

	var req updateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	task, err := h.service.Update(
		r.Context(),
		UpdateTaskInput{
			TaskID:         taskID,
			UserID:         userID,
			QueueID:        req.QueueID,
			Name:           req.Name,
			Description:    req.Description,
			Status:         req.Status,
			TimeoutSeconds: req.TimeoutSeconds,
			MaxRetries:     req.MaxRetries,
		},
	)

	if errors.Is(err, ErrTaskNotFound) {
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"TASK_NOT_FOUND",
			"task not found",
		)
		return
	}

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"TASK_UPDATE_FAILED",
			err.Error(),
		)
		return
	}

	response.JSON(w, http.StatusOK, task)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())

	if !ok || strings.TrimSpace(userID) == "" {
		response.ErrorJSON(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"user authentication is required",
		)
		return
	}

	taskID := r.PathValue("id")

	err := h.service.Archive(r.Context(), taskID, userID)

	if errors.Is(err, ErrTaskNotFound) {
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"TASK_NOT_FOUND",
			"task not found",
		)
		return
	}

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"TASK_ARCHIVE_FAILED",
			err.Error(),
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
