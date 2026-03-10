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

func writeInternalError(w http.ResponseWriter) {
	var resp errorResponse
	resp.Error.Message = "internal server error"
	writeJSON(w, http.StatusInternalServerError, resp)
}
