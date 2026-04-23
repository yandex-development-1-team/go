//go:build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	pgrepo "github.com/yandex-development-1-team/go/internal/repository/postgres"
	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

func generateAdminToken(t *testing.T) string {
	t.Helper()
	claims := svcapi.AccessClaims{
		UserID: 1,
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)
	return signed
}

func setupUsersAdminServer(t *testing.T) *httptest.Server {
	t.Helper()

	staffRepo := pgrepo.NewStaffRepo(db)
	refreshRepo := pgrepo.NewRefreshTokenRepo(db)
	svc := svcapi.NewUsersAdminService(staffRepo, refreshRepo)
	handler := NewUsersHandler(svc)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": []string{"unauthorized"}})
			return
		}
		tokenStr := authHeader[len("Bearer "):]
		claims := &svcapi.AccessClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": []string{"unauthorized"}})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	})

	users := router.Group("/users")
	users.Use(func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"errors": []string{"forbidden"}})
			return
		}
		c.Next()
	})
	{
		users.POST("", handler.Create)
		users.PUT("/:id", handler.Update)
		users.PUT("/:id/block", handler.Block)
	}

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server
}

func authHeader(token string) string {
	return "Bearer " + token
}

func ptr(s string) *string { return &s }

func TestUsersAdmin_Create_Success(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	body := map[string]interface{}{
		"first_name": "Иван",
		"last_name":  "Иванов",
		"email":      "ivan@example.com",
		"role":       "manager_1",
	}
	b, _ := json.Marshal(body)

	resp, err := doRequest(server.URL+"/users", http.MethodPost, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "ivan@example.com", result["email"])
	assert.Equal(t, "manager_1", result["role"])
	assert.Equal(t, "invited", result["status"])
}

func TestUsersAdmin_Create_AllFields(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	body := map[string]interface{}{
		"first_name":  "Мария",
		"last_name":   "Петрова",
		"second_name": "Сергеевна",
		"email":       "maria@example.com",
		"role":        "manager_2",
		"department":  "Маркетинг",
		"position":    "Менеджер",
		"supervisor":  "Иван Иванов",
		"address":     "Москва",
	}
	b, _ := json.Marshal(body)

	resp, err := doRequest(server.URL+"/users", http.MethodPost, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "maria@example.com", result["email"])
	assert.Equal(t, "manager_2", result["role"])
}

func TestUsersAdmin_Create_DuplicateEmail(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	body := map[string]interface{}{
		"first_name": "Иван",
		"last_name":  "Иванов",
		"email":      "dup@example.com",
		"role":       "manager_1",
	}
	b, _ := json.Marshal(body)

	resp1, _ := doRequest(server.URL+"/users", http.MethodPost, b, authHeader(token))
	resp1.Body.Close()

	resp2, err := doRequest(server.URL+"/users", http.MethodPost, b, authHeader(token))
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
}

func TestUsersAdmin_Create_InvalidRole(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	body := map[string]interface{}{
		"first_name": "Иван",
		"last_name":  "Иванов",
		"email":      "role@example.com",
		"role":       "superadmin",
	}
	b, _ := json.Marshal(body)

	resp, err := doRequest(server.URL+"/users", http.MethodPost, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUsersAdmin_Create_MissingRequiredFields(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	body := map[string]interface{}{
		"first_name": "Иван",
		"last_name":  "Иванов",
		"role":       "manager_1",
	}
	b, _ := json.Marshal(body)

	resp, err := doRequest(server.URL+"/users", http.MethodPost, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUsersAdmin_Create_Unauthorized(t *testing.T) {
	server := setupUsersAdminServer(t)

	body := map[string]interface{}{
		"first_name": "Иван",
		"last_name":  "Иванов",
		"email":      "unauth@example.com",
		"role":       "manager_1",
	}
	b, _ := json.Marshal(body)

	resp, err := doRequest(server.URL+"/users", http.MethodPost, b, "")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUsersAdmin_Update_Success(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	id := seedStaffAdmin(t, "update@example.com")

	body := map[string]interface{}{
		"first_name": "Обновлённое",
		"department": "IT",
		"position":   "Senior",
		"supervisor": "Начальник",
		"address":    "Санкт-Петербург",
	}
	b, _ := json.Marshal(body)

	resp, err := doRequest(fmt.Sprintf("%s/users/%d", server.URL, id), http.MethodPut, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "Обновлённое", result["name"])
}

func TestUsersAdmin_Update_NotFound(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	body := map[string]interface{}{"first_name": "Тест"}
	b, _ := json.Marshal(body)

	resp, err := doRequest(server.URL+"/users/99999", http.MethodPut, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUsersAdmin_Update_InvalidID(t *testing.T) {
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	body := map[string]interface{}{"first_name": "Тест"}
	b, _ := json.Marshal(body)

	resp, err := doRequest(server.URL+"/users/abc", http.MethodPut, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUsersAdmin_Update_InvalidRole(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	id := seedStaffAdmin(t, "role_upd@example.com")

	role := "superadmin"
	body := map[string]interface{}{"role": role}
	b, _ := json.Marshal(body)

	resp, err := doRequest(fmt.Sprintf("%s/users/%d", server.URL, id), http.MethodPut, b, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUsersAdmin_Block_Success(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	id := seedStaffAdmin(t, "block@example.com")

	resp, err := doRequest(fmt.Sprintf("%s/users/%d/block", server.URL, id), http.MethodPut, nil, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "blocked", result["status"])
	assert.Equal(t, float64(id), result["id"])
}

func TestUsersAdmin_Block_UserCannotLoginAfterBlock(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	var id int64
	err := db.QueryRow(`
		INSERT INTO staff (first_name, last_name, email, password_hash, role, status)
		VALUES ('Тест', 'Блок', 'loginblock@example.com', $1, 'manager_1', 'active')
		RETURNING id`, string(hash)).Scan(&id)
	require.NoError(t, err)

	resp, err := doRequest(fmt.Sprintf("%s/users/%d/block", server.URL, id), http.MethodPut, nil, authHeader(token))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var status string
	err = db.QueryRow(`SELECT status FROM staff WHERE id = $1`, id).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "blocked", status)
}

func TestUsersAdmin_Block_RefreshTokensRevoked(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	_, _ = db.Exec(`TRUNCATE TABLE refresh_tokens CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	id := seedStaffAdmin(t, "revoke@example.com")

	_, err := db.Exec(`
    	INSERT INTO refresh_tokens (user_id, token, expires_at)
    	VALUES ($1, 'test-refresh-token', NOW() + INTERVAL '7 days')`, id)
	require.NoError(t, err)

	var count int
	_ = db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1`, id).Scan(&count)
	assert.Equal(t, 1, count)

	resp, err := doRequest(fmt.Sprintf("%s/users/%d/block", server.URL, id), http.MethodPut, nil, authHeader(token))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	_ = db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1`, id).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestUsersAdmin_Block_NotFound(t *testing.T) {
	_, _ = db.Exec(`TRUNCATE TABLE staff CASCADE`)
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	resp, err := doRequest(server.URL+"/users/99999/block", http.MethodPut, nil, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUsersAdmin_Block_InvalidID(t *testing.T) {
	server := setupUsersAdminServer(t)
	token := generateAdminToken(t)

	resp, err := doRequest(server.URL+"/users/abc/block", http.MethodPut, nil, authHeader(token))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func seedStaffAdmin(t *testing.T, email string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(`
		INSERT INTO staff (first_name, last_name, email, role, status)
		VALUES ('Тест', 'Тестов', $1, 'manager_1', 'active')
		RETURNING id`, email).Scan(&id)
	require.NoError(t, err)
	return id
}

func doRequest(url, method string, body []byte, auth string) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	return http.DefaultClient.Do(req)
}
