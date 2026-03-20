package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
)

type applicationHandler struct {
	repo repository.ApplicationRepository
}

func NewApplicationHandler(repo repository.ApplicationRepository) *applicationHandler {
	return &applicationHandler{repo: repo}
}

func (h *applicationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/applications", h.create)
	mux.HandleFunc("GET /api/v1/applications", h.list)
	mux.HandleFunc("GET /api/v1/applications/{id}", h.getByID)
	mux.HandleFunc("PUT /api/v1/applications/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/applications/{id}", h.delete)
}

func (h *applicationHandler) create(w http.ResponseWriter, r *http.Request) {
	var req models.ApplicationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, validationDetail{Field: "body", Message: "invalid JSON"})
		return
	}

	var details []validationDetail
	if !req.Type.Valid() {
		details = append(details, validationDetail{Field: "type", Message: "must be one of: box, special_project"})
	}
	if !req.Source.Valid() {
		details = append(details, validationDetail{Field: "source", Message: "must be one of: telegram_bot, manual"})
	}
	if req.CustomerName == "" {
		details = append(details, validationDetail{Field: "customer_name", Message: "required"})
	}
	if req.ContactInfo == "" {
		details = append(details, validationDetail{Field: "contact_info", Message: "required"})
	}
	if len(details) > 0 {
		writeValidationErrors(w, details)
		return
	}

	app, err := h.repo.CreateApplication(r.Context(), &req)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			writeValidationError(w, validationDetail{Field: "body", Message: err.Error()})
			return
		}
		logger.Error("failed to create application", zap.Error(err))
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusCreated, app)
}

func (h *applicationHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := models.ApplicationFilter{Limit: 20}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			filter.Offset = n
		}
	}
	if v := q.Get("status"); v != "" {
		s := models.ApplicationStatus(v)
		filter.Status = &s
	}
	if v := q.Get("type"); v != "" {
		t := models.ApplicationType(v)
		filter.Type = &t
	}
	if v := q.Get("manager_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.ManagerID = &id
		}
	}
	if v := q.Get("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.DateFrom = &t
		}
	}
	if v := q.Get("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.DateTo = &t
		}
	}

	apps, total, err := h.repo.GetApplications(r.Context(), filter)
	if err != nil {
		logger.Error("failed to get applications", zap.Error(err))
		writeInternalError(w)
		return
	}

	if apps == nil {
		apps = []models.Application{}
	}

	writeJSON(w, http.StatusOK, models.ApplicationListResponse{
		Items:      apps,
		Pagination: models.Pagination{Limit: filter.Limit, Offset: filter.Offset, Total: total},
	})
}

func (h *applicationHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	app, err := h.repo.GetApplicationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrApplicationNotFound) {
			writeNotFoundError(w)
			return
		}
		logger.Error("failed to get application by id", zap.Error(err))
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (h *applicationHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req models.ApplicationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, validationDetail{Field: "body", Message: "invalid JSON"})
		return
	}

	if req.Status == nil && req.ContactInfo == nil && req.BoxID == nil && req.SpecialProjectID == nil {
		writeValidationError(w, validationDetail{Field: "body", Message: "at least one field must be provided"})
		return
	}

	if req.Status != nil && !req.Status.Valid() {
		writeValidationError(w, validationDetail{Field: "status", Message: "must be one of: queue, in_progress, done"})
		return
	}

	app, err := h.repo.UpdateApplication(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, models.ErrApplicationNotFound) {
			writeNotFoundError(w)
			return
		}
		if errors.Is(err, models.ErrInvalidInput) {
			writeValidationError(w, validationDetail{Field: "body", Message: err.Error()})
			return
		}
		logger.Error("failed to update application", zap.Error(err))
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (h *applicationHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.repo.DeleteApplication(r.Context(), id); err != nil {
		if errors.Is(err, models.ErrApplicationNotFound) {
			writeNotFoundError(w)
			return
		}
		logger.Error("failed to delete application", zap.Error(err))
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "application deleted"})
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		writeValidationError(w, validationDetail{Field: "id", Message: "must be a positive integer"})
		return 0, false
	}
	return id, true
}

type validationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error struct {
		Details []validationDetail `json:"details,omitempty"`
		Message string             `json:"message,omitempty"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Error("failed to encode response", zap.Error(err))
	}
}

func writeValidationError(w http.ResponseWriter, detail validationDetail) {
	writeValidationErrors(w, []validationDetail{detail})
}

func writeValidationErrors(w http.ResponseWriter, details []validationDetail) {
	var resp errorResponse
	resp.Error.Details = details
	writeJSON(w, http.StatusBadRequest, resp)
}

func writeNotFoundError(w http.ResponseWriter) {
	var resp errorResponse
	resp.Error.Message = "not found"
	writeJSON(w, http.StatusNotFound, resp)
}

func writeInternalError(w http.ResponseWriter) {
	var resp errorResponse
	resp.Error.Message = "internal server error"
	writeJSON(w, http.StatusInternalServerError, resp)
}
