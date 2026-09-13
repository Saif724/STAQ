package queues

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

type createQueueRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type updateQueueRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createQueueRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	queue, err := h.service.Create(
		r.Context(),
		CreateQueueInput{
			Name:        req.Name,
			Description: req.Description,
		},
	)

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"QUEUE_CREATION_FAILED",
			err.Error(),
		)
		return
	}

	response.JSON(
		w,
		http.StatusCreated,
		queue,
	)
}

func (h *Handler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	queues, err := h.service.List(r.Context())
	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"QUEUE_LIST_FAILED",
			err.Error(),
		)
		return
	}

	response.JSON(w, http.StatusOK, queues)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	queueID := r.PathValue("id")

	if strings.TrimSpace(queueID) == "" {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_QUEUE_ID",
			"queue id is required",
		)
		return
	}

	queue, err := h.service.GetByID(
		r.Context(),
		queueID,
	)

	if errors.Is(err, ErrQueueNotFound) {
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"QUEUE_NOT_FOUND",
			"queue not found",
		)
		return
	}

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusInternalServerError,
			"QUEUE_FETCH_FAILED",
			err.Error(),
		)
		return
	}

	response.JSON(w, http.StatusOK, queue)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	queueID := r.PathValue("id")

	if strings.TrimSpace(queueID) == "" {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_QUEUE_ID",
			"queue id is required",
		)
		return
	}

	var req updateQueueRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	queue, err := h.service.Update(
		r.Context(),
		UpdateQueueInput{
			ID:          queueID,
			Name:        req.Name,
			Description: req.Description,
			IsActive:    req.IsActive,
		},
	)

	if errors.Is(err, ErrQueueNotFound) {
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"QUEUE_NOT_FOUND",
			"queue not found",
		)
		return
	}

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"QUEUE_UPDATE_FAILED",
			err.Error(),
		)

		return
	}

	response.JSON(w, http.StatusOK, queue)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	queueID := r.PathValue("id")

	if strings.TrimSpace(queueID) == "" {
		response.ErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_QUEUE_ID",
			"queue id is required",
		)
		return
	}

	err := h.service.Delete(
		r.Context(),
		queueID,
	)

	if errors.Is(err, ErrQueueNotFound) {
		response.ErrorJSON(
			w,
			http.StatusNotFound,
			"QUEUE_NOT_FOUND",
			"queue not found",
		)
		return
	}

	if err != nil {
		response.ErrorJSON(
			w,
			http.StatusConflict,
			"QUEUE_DELETE_FAILED",
			"queue cannot be deleted because it may be in use",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
