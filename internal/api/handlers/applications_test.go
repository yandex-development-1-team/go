package handlers

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"github.com/yandex-development-1-team/go/internal/apierrors"
// 	"github.com/yandex-development-1-team/go/internal/dto"
// 	"github.com/yandex-development-1-team/go/internal/models"
// )

// type mockApplicationRepo struct {
// 	createFn  func(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error)
// 	listFn    func(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error)
// 	getByIDFn func(ctx context.Context, id int64) (*models.Application, error)
// 	updateFn  func(ctx context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error)
// 	deleteFn  func(ctx context.Context, id int64) error
// }

// func (m *mockApplicationRepo) CreateApplication(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error) {
// 	if m.createFn == nil {
// 		return nil, errors.New("unexpected call to CreateApplication")
// 	}
// 	return m.createFn(ctx, req)
// }

// func (m *mockApplicationRepo) GetApplications(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 	if m.listFn == nil {
// 		return nil, 0, errors.New("unexpected call to GetApplications")
// 	}
// 	return m.listFn(ctx, filter)
// }

// func (m *mockApplicationRepo) GetApplicationByID(ctx context.Context, id int64) (*models.Application, error) {
// 	if m.getByIDFn == nil {
// 		return nil, errors.New("unexpected call to GetApplicationByID")
// 	}
// 	return m.getByIDFn(ctx, id)
// }

// func (m *mockApplicationRepo) UpdateApplication(ctx context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error) {
// 	if m.updateFn == nil {
// 		return nil, errors.New("unexpected call to UpdateApplication")
// 	}
// 	return m.updateFn(ctx, id, req)
// }

// func (m *mockApplicationRepo) DeleteApplication(ctx context.Context, id int64) error {
// 	if m.deleteFn == nil {
// 		return errors.New("unexpected call to DeleteApplication")
// 	}
// 	return m.deleteFn(ctx, id)
// }

// func newApplicationTestRouter(repo *mockApplicationRepo) *gin.Engine {
// 	gin.SetMode(gin.TestMode)

// 	r := gin.New()
// 	h := NewApplicationHandler(repo)

// 	r.POST("/applications", h.Create)
// 	r.GET("/applications", h.List)
// 	r.GET("/applications/:id", h.GetByID)
// 	r.PUT("/applications/:id", h.Update)
// 	r.DELETE("/applications/:id", h.Delete)

// 	return r
// }

// func performJSONRequest(
// 	t *testing.T,
// 	router http.Handler,
// 	method string,
// 	url string,
// 	body any,
// ) *httptest.ResponseRecorder {
// 	t.Helper()

// 	var reader *bytes.Reader
// 	if body == nil {
// 		reader = bytes.NewReader(nil)
// 	} else {
// 		raw, err := json.Marshal(body)
// 		require.NoError(t, err)
// 		reader = bytes.NewReader(raw)
// 	}

// 	req := httptest.NewRequest(method, url, reader)
// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	return w
// }

// func performRawRequest(
// 	t *testing.T,
// 	router http.Handler,
// 	method string,
// 	url string,
// 	raw string,
// ) *httptest.ResponseRecorder {
// 	t.Helper()

// 	req := httptest.NewRequest(method, url, bytes.NewBufferString(raw))
// 	req.Header.Set("Content-Type", "application/json")

// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
// 	return w
// }

// func sampleApplication() *models.Application {
// 	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)

// 	managerID := int64(7)
// 	boxID := int64(11)
// 	managerName := "Анна Петрова"
// 	projectName := "Проект А"

// 	return &models.Application{
// 		ID:           42,
// 		Type:         models.ApplicationTypeBox,
// 		Source:       models.ApplicationSourceManual,
// 		Status:       models.ApplicationStatusQueue,
// 		CustomerName: "Иван Иванов",
// 		ContactInfo:  "ivan@example.com",
// 		ProjectName:  &projectName,
// 		BoxID:        &boxID,
// 		ManagerID:    &managerID,
// 		ManagerName:  &managerName,
// 		CreatedAt:    now,
// 		UpdatedAt:    now,
// 	}
// }

// func assertAPIErrorsBody(t *testing.T, body []byte) apierrors.ServiceErrorResponse {
// 	t.Helper()

// 	var resp apierrors.ServiceErrorResponse
// 	require.NoError(t, json.Unmarshal(body, &resp))
// 	require.NotNil(t, resp.Errors)
// 	require.NotEmpty(t, resp.Errors)

// 	return resp
// }

// func assertAPIErrorsBodyHasExactMessage(t *testing.T, body []byte, want string) {
// 	t.Helper()

// 	resp := assertAPIErrorsBody(t, body)
// 	assert.Contains(t, resp.Errors, want)
// }

// func TestApplicationHandler_Create(t *testing.T) {
// 	t.Run("success", func(t *testing.T) {
// 		var captured *models.ApplicationCreateRequest

// 		repo := &mockApplicationRepo{
// 			createFn: func(_ context.Context, req *models.ApplicationCreateRequest) (*models.Application, error) {
// 				captured = req
// 				return sampleApplication(), nil
// 			},
// 		}

// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPost, "/applications", models.ApplicationCreateRequest{
// 			Type:         models.ApplicationTypeBox,
// 			Source:       models.ApplicationSourceManual,
// 			CustomerName: "Иван Иванов",
// 			ContactInfo:  "ivan@example.com",
// 		})

// 		require.Equal(t, http.StatusCreated, w.Code)

// 		require.NotNil(t, captured)
// 		assert.Equal(t, models.ApplicationTypeBox, captured.Type)
// 		assert.Equal(t, models.ApplicationSourceManual, captured.Source)
// 		assert.Equal(t, "Иван Иванов", captured.CustomerName)
// 		assert.Equal(t, "ivan@example.com", captured.ContactInfo)

// 		var resp models.Application
// 		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
// 		assert.Equal(t, int64(42), resp.ID)
// 		assert.Equal(t, "Иван Иванов", resp.CustomerName)
// 		assert.Contains(t, w.Body.String(), `"manager_name"`)
// 	})

// 	t.Run("invalid json", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performRawRequest(t, router, http.MethodPost, "/applications", `{"type":`)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверный формат запроса")
// 	})

// 	t.Run("missing customer_name", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPost, "/applications", models.ApplicationCreateRequest{
// 			Type:        models.ApplicationTypeBox,
// 			Source:      models.ApplicationSourceManual,
// 			ContactInfo: "ivan@example.com",
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректные данные заявки")
// 	})

// 	t.Run("missing contact_info", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPost, "/applications", models.ApplicationCreateRequest{
// 			Type:         models.ApplicationTypeBox,
// 			Source:       models.ApplicationSourceManual,
// 			CustomerName: "Иван Иванов",
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректные данные заявки")
// 	})

// 	t.Run("invalid type", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPost, "/applications", models.ApplicationCreateRequest{
// 			Type:         models.ApplicationType("invalid"),
// 			Source:       models.ApplicationSourceManual,
// 			CustomerName: "Иван Иванов",
// 			ContactInfo:  "ivan@example.com",
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректные данные заявки")
// 	})

// 	t.Run("invalid source", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPost, "/applications", models.ApplicationCreateRequest{
// 			Type:         models.ApplicationTypeBox,
// 			Source:       models.ApplicationSource("invalid"),
// 			CustomerName: "Иван Иванов",
// 			ContactInfo:  "ivan@example.com",
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректные данные заявки")
// 	})

// 	t.Run("repo invalid input", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			createFn: func(_ context.Context, _ *models.ApplicationCreateRequest) (*models.Application, error) {
// 				return nil, models.ErrInvalidInput
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPost, "/applications", models.ApplicationCreateRequest{
// 			Type:         models.ApplicationTypeBox,
// 			Source:       models.ApplicationSourceManual,
// 			CustomerName: "Иван Иванов",
// 			ContactInfo:  "ivan@example.com",
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})

// 	t.Run("internal error", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			createFn: func(_ context.Context, _ *models.ApplicationCreateRequest) (*models.Application, error) {
// 				return nil, errors.New("db exploded")
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPost, "/applications", models.ApplicationCreateRequest{
// 			Type:         models.ApplicationTypeBox,
// 			Source:       models.ApplicationSourceManual,
// 			CustomerName: "Иван Иванов",
// 			ContactInfo:  "ivan@example.com",
// 		})

// 		require.Equal(t, http.StatusInternalServerError, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})
// }

// func TestApplicationHandler_List(t *testing.T) {
// 	t.Run("success", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 				assert.Equal(t, dto.DefaultApplicationLimit, filter.Limit)
// 				assert.Equal(t, 0, filter.Offset)
// 				return []models.Application{*sampleApplication()}, 1, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications", nil)

// 		require.Equal(t, http.StatusOK, w.Code)

// 		var resp dto.ApplicationListResponse
// 		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
// 		require.Len(t, resp.Items, 1)
// 		assert.Equal(t, 1, resp.Pagination.Total)
// 		assert.Equal(t, dto.DefaultApplicationLimit, resp.Pagination.Limit)
// 		assert.Equal(t, 0, resp.Pagination.Offset)
// 		assert.Contains(t, w.Body.String(), `"manager_name"`)
// 	})

// 	t.Run("empty returns empty array", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, _ models.ApplicationFilter) ([]models.Application, int, error) {
// 				return nil, 0, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications", nil)

// 		require.Equal(t, http.StatusOK, w.Code)

// 		var resp dto.ApplicationListResponse
// 		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
// 		require.NotNil(t, resp.Items)
// 		assert.Len(t, resp.Items, 0)
// 		assert.Equal(t, 0, resp.Pagination.Total)
// 	})

// 	t.Run("default pagination", func(t *testing.T) {
// 		var captured models.ApplicationFilter

// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 				captured = filter
// 				return []models.Application{}, 0, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		assert.Equal(t, dto.DefaultApplicationLimit, captured.Limit)
// 		assert.Equal(t, 0, captured.Offset)
// 	})

// 	t.Run("custom pagination", func(t *testing.T) {
// 		var captured models.ApplicationFilter

// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 				captured = filter
// 				return []models.Application{}, 0, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?limit=50&offset=10", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		assert.Equal(t, 50, captured.Limit)
// 		assert.Equal(t, 10, captured.Offset)
// 	})

// 	t.Run("invalid limit string", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?limit=abc", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("invalid limit zero", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?limit=0", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("limit too large", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?limit=101", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("invalid offset string", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?offset=abc", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("invalid negative offset", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?offset=-1", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("filter by customer_name", func(t *testing.T) {
// 		var captured models.ApplicationFilter

// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 				captured = filter
// 				return []models.Application{}, 0, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?customer_name=Иван", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		assert.Equal(t, "Иван", captured.CustomerName)
// 	})

// 	t.Run("filter by type", func(t *testing.T) {
// 		var captured models.ApplicationFilter

// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 				captured = filter
// 				return []models.Application{}, 0, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?type=box", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		require.NotNil(t, captured.Type)
// 		assert.Equal(t, models.ApplicationTypeBox, *captured.Type)
// 	})

// 	t.Run("filter by status", func(t *testing.T) {
// 		var captured models.ApplicationFilter

// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 				captured = filter
// 				return []models.Application{}, 0, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?status=queue", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		require.NotNil(t, captured.Status)
// 		assert.Equal(t, models.ApplicationStatusQueue, *captured.Status)
// 	})

// 	t.Run("filter by manager_id", func(t *testing.T) {
// 		var captured models.ApplicationFilter

// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
// 				captured = filter
// 				return []models.Application{}, 0, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?manager_id=42", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		require.NotNil(t, captured.ManagerID)
// 		assert.Equal(t, int64(42), *captured.ManagerID)
// 	})

// 	t.Run("invalid manager_id", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?manager_id=abc", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("response contains manager_name", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, _ models.ApplicationFilter) ([]models.Application, int, error) {
// 				return []models.Application{*sampleApplication()}, 1, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		assert.Contains(t, w.Body.String(), `"manager_name":"Анна Петрова"`)
// 	})

// 	t.Run("date_from is not supported", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?date_from=2026-04-01T00:00:00Z", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("date_to is not supported", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?date_to=2026-04-01T00:00:00Z", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверные параметры запроса")
// 	})

// 	t.Run("internal error", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, _ models.ApplicationFilter) ([]models.Application, int, error) {
// 				return nil, 0, errors.New("db exploded")
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications", nil)

// 		require.Equal(t, http.StatusInternalServerError, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})
// }

// func TestApplicationHandler_GetByID(t *testing.T) {
// 	t.Run("success", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			getByIDFn: func(_ context.Context, id int64) (*models.Application, error) {
// 				app := sampleApplication()
// 				app.ID = id
// 				return app, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications/42", nil)

// 		require.Equal(t, http.StatusOK, w.Code)

// 		var resp models.Application
// 		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
// 		assert.Equal(t, int64(42), resp.ID)
// 		assert.Contains(t, w.Body.String(), `"manager_name"`)
// 	})

// 	t.Run("invalid id alpha", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications/abc", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректный идентификатор заявки")
// 	})

// 	t.Run("invalid id zero", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications/0", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректный идентификатор заявки")
// 	})

// 	t.Run("not found", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			getByIDFn: func(_ context.Context, _ int64) (*models.Application, error) {
// 				return nil, models.ErrApplicationNotFound
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications/999", nil)

// 		require.Equal(t, http.StatusNotFound, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})

// 	t.Run("internal error", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			getByIDFn: func(_ context.Context, _ int64) (*models.Application, error) {
// 				return nil, errors.New("db exploded")
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications/42", nil)

// 		require.Equal(t, http.StatusInternalServerError, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})
// }

// func TestApplicationHandler_Update(t *testing.T) {
// 	t.Run("success", func(t *testing.T) {
// 		status := models.ApplicationStatusInProgress
// 		contactInfo := "new@example.com"

// 		var capturedID int64
// 		var capturedReq *models.ApplicationUpdateRequest

// 		repo := &mockApplicationRepo{
// 			updateFn: func(_ context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error) {
// 				capturedID = id
// 				capturedReq = req

// 				app := sampleApplication()
// 				app.Status = status
// 				app.ContactInfo = contactInfo
// 				return app, nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPut, "/applications/42", models.ApplicationUpdateRequest{
// 			Status:      &status,
// 			ContactInfo: &contactInfo,
// 		})

// 		require.Equal(t, http.StatusOK, w.Code)
// 		assert.Equal(t, int64(42), capturedID)
// 		require.NotNil(t, capturedReq)
// 		require.NotNil(t, capturedReq.Status)
// 		require.NotNil(t, capturedReq.ContactInfo)
// 		assert.Equal(t, models.ApplicationStatusInProgress, *capturedReq.Status)
// 		assert.Equal(t, "new@example.com", *capturedReq.ContactInfo)
// 		assert.Contains(t, w.Body.String(), `"manager_name"`)
// 	})

// 	t.Run("invalid json", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performRawRequest(t, router, http.MethodPut, "/applications/42", `{"status":`)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Неверный формат запроса")
// 	})

// 	t.Run("invalid id", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		status := models.ApplicationStatusDone
// 		w := performJSONRequest(t, router, http.MethodPut, "/applications/abc", models.ApplicationUpdateRequest{
// 			Status: &status,
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректный идентификатор заявки")
// 	})

// 	t.Run("no fields", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPut, "/applications/42", models.ApplicationUpdateRequest{})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Нужно передать хотя бы одно поле для обновления")
// 	})

// 	t.Run("invalid status", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		invalidStatus := models.ApplicationStatus("unknown")
// 		w := performJSONRequest(t, router, http.MethodPut, "/applications/42", models.ApplicationUpdateRequest{
// 			Status: &invalidStatus,
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректный статус заявки")
// 	})

// 	t.Run("not found", func(t *testing.T) {
// 		status := models.ApplicationStatusDone

// 		repo := &mockApplicationRepo{
// 			updateFn: func(_ context.Context, _ int64, _ *models.ApplicationUpdateRequest) (*models.Application, error) {
// 				return nil, models.ErrApplicationNotFound
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPut, "/applications/999", models.ApplicationUpdateRequest{
// 			Status: &status,
// 		})

// 		require.Equal(t, http.StatusNotFound, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})

// 	t.Run("repo invalid input", func(t *testing.T) {
// 		status := models.ApplicationStatusDone

// 		repo := &mockApplicationRepo{
// 			updateFn: func(_ context.Context, _ int64, _ *models.ApplicationUpdateRequest) (*models.Application, error) {
// 				return nil, models.ErrInvalidInput
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPut, "/applications/42", models.ApplicationUpdateRequest{
// 			Status: &status,
// 		})

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})

// 	t.Run("internal error", func(t *testing.T) {
// 		status := models.ApplicationStatusDone

// 		repo := &mockApplicationRepo{
// 			updateFn: func(_ context.Context, _ int64, _ *models.ApplicationUpdateRequest) (*models.Application, error) {
// 				return nil, errors.New("db exploded")
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodPut, "/applications/42", models.ApplicationUpdateRequest{
// 			Status: &status,
// 		})

// 		require.Equal(t, http.StatusInternalServerError, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})
// }

// func TestApplicationHandler_Delete(t *testing.T) {
// 	t.Run("success", func(t *testing.T) {
// 		var capturedID int64

// 		repo := &mockApplicationRepo{
// 			deleteFn: func(_ context.Context, id int64) error {
// 				capturedID = id
// 				return nil
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodDelete, "/applications/42", nil)

// 		require.Equal(t, http.StatusOK, w.Code)
// 		assert.Equal(t, int64(42), capturedID)
// 		assert.Contains(t, w.Body.String(), "message")
// 	})

// 	t.Run("invalid id", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodDelete, "/applications/abc", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)
// 		assertAPIErrorsBodyHasExactMessage(t, w.Body.Bytes(), "Некорректный идентификатор заявки")
// 	})

// 	t.Run("not found", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			deleteFn: func(_ context.Context, _ int64) error {
// 				return models.ErrApplicationNotFound
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodDelete, "/applications/999", nil)

// 		require.Equal(t, http.StatusNotFound, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})

// 	t.Run("internal error", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			deleteFn: func(_ context.Context, _ int64) error {
// 				return errors.New("db exploded")
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodDelete, "/applications/42", nil)

// 		require.Equal(t, http.StatusInternalServerError, w.Code)
// 		assertAPIErrorsBody(t, w.Body.Bytes())
// 	})
// }

// func TestApplicationHandler_ErrorFormat(t *testing.T) {
// 	t.Run("bad request uses apierrors", func(t *testing.T) {
// 		repo := &mockApplicationRepo{}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications?limit=abc", nil)

// 		require.Equal(t, http.StatusBadRequest, w.Code)

// 		var resp apierrors.ServiceErrorResponse
// 		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
// 		require.NotNil(t, resp.Errors)
// 		require.NotEmpty(t, resp.Errors)
// 	})

// 	t.Run("not found uses apierrors", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			getByIDFn: func(_ context.Context, _ int64) (*models.Application, error) {
// 				return nil, models.ErrApplicationNotFound
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications/404", nil)

// 		require.Equal(t, http.StatusNotFound, w.Code)

// 		var resp apierrors.ServiceErrorResponse
// 		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
// 		require.NotEmpty(t, resp.Errors)
// 	})

// 	t.Run("internal error uses apierrors", func(t *testing.T) {
// 		repo := &mockApplicationRepo{
// 			listFn: func(_ context.Context, _ models.ApplicationFilter) ([]models.Application, int, error) {
// 				return nil, 0, errors.New("boom")
// 			},
// 		}
// 		router := newApplicationTestRouter(repo)

// 		w := performJSONRequest(t, router, http.MethodGet, "/applications", nil)

// 		require.Equal(t, http.StatusInternalServerError, w.Code)

// 		var resp apierrors.ServiceErrorResponse
// 		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
// 		require.NotEmpty(t, resp.Errors)
// 	})
// }
