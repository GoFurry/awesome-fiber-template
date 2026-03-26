package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/medium/config"
	"github.com/GoFurry/awesome-go-template/fiber/v3/medium/internal/bootstrap"
	usermodels "github.com/GoFurry/awesome-go-template/fiber/v3/medium/internal/app/user/models"
	"github.com/GoFurry/awesome-go-template/fiber/v3/medium/internal/infra/db"
	"github.com/GoFurry/awesome-go-template/fiber/v3/medium/internal/transport/http/router"
	"github.com/GoFurry/awesome-go-template/fiber/v3/medium/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type apiResponse struct {
	Status  int             `json:"-"`
	Headers http.Header     `json:"-"`
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type userResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Status string `json:"status"`
}

type userListResponse struct {
	Total int64          `json:"total"`
	List  []userResponse `json:"list"`
}

func TestBootstrapWithSQLiteCRUD(t *testing.T) {
	configPath, databasePath := writeIntegrationConfig(t)
	env.MustInitServerConfig(common.COMMON_PROJECT_NAME, configPath)

	if err := bootstrap.Start(); err != nil {
		t.Fatalf("start bootstrap failed: %v", err)
	}

	httpApp := router.New().Init()
	t.Cleanup(func() {
		_ = httpApp.Shutdown()
		_ = bootstrap.Shutdown()
	})

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("sqlite database file not created: %v", err)
	}
	if !db.Orm.DB().Migrator().HasTable(&usermodels.User{}) {
		t.Fatalf("expected users table to be auto migrated")
	}

	health := doRequest(t, httpApp, http.MethodGet, "/healthz", nil)
	if health.Status != http.StatusOK {
		t.Fatalf("healthz returned unexpected status: %+v", health)
	}
	if health.Code != common.RETURN_SUCCESS {
		t.Fatalf("healthz returned unexpected code: %+v", health)
	}
	if health.Headers.Get(fiber.HeaderXRequestID) == "" {
		t.Fatalf("expected request id header on healthz response")
	}
	if health.Headers.Get(fiber.HeaderXContentTypeOptions) == "" {
		t.Fatalf("expected security headers on healthz response")
	}

	livez := rawRequest(t, httpApp, http.MethodGet, "/livez", nil, nil)
	defer livez.Body.Close()
	if livez.StatusCode != http.StatusOK {
		t.Fatalf("livez returned unexpected status: %d", livez.StatusCode)
	}

	readyz := rawRequest(t, httpApp, http.MethodGet, "/readyz", nil, nil)
	defer readyz.Body.Close()
	if readyz.StatusCode != http.StatusOK {
		t.Fatalf("readyz returned unexpected status: %d", readyz.StatusCode)
	}

	startupz := rawRequest(t, httpApp, http.MethodGet, "/startupz", nil, nil)
	defer startupz.Body.Close()
	if startupz.StatusCode != http.StatusOK {
		t.Fatalf("startupz returned unexpected status: %d", startupz.StatusCode)
	}

	createBody := map[string]any{
		"name":   "Alice",
		"email":  "alice@example.com",
		"age":    28,
		"status": "active",
	}
	create := doRequest(t, httpApp, http.MethodPost, "/api/v1/user/", createBody)
	var created userResponse
	mustDecode(t, create.Data, &created)
	if created.ID == 0 {
		t.Fatalf("expected created user id, got %+v", created)
	}

	list := doRequest(t, httpApp, http.MethodGet, "/api/v1/user/?page_num=1&page_size=10", nil)
	if list.Headers.Get(fiber.HeaderETag) == "" {
		t.Fatalf("expected ETag header on list response")
	}
	var users userListResponse
	mustDecode(t, list.Data, &users)
	if users.Total < 1 || len(users.List) < 1 {
		t.Fatalf("unexpected list payload: %+v", users)
	}

	compressed := rawRequest(t, httpApp, http.MethodGet, "/api/v1/user/?page_num=1&page_size=10", nil, map[string]string{
		fiber.HeaderAcceptEncoding: "gzip",
	})
	defer compressed.Body.Close()
	if compressed.Header.Get(fiber.HeaderContentEncoding) == "" {
		t.Fatalf("expected compressed response when Accept-Encoding is set")
	}

	get := doRequest(t, httpApp, http.MethodGet, "/api/v1/user/"+toString(created.ID), nil)
	var fetched userResponse
	mustDecode(t, get.Data, &fetched)
	if fetched.Email != "alice@example.com" {
		t.Fatalf("unexpected fetched user: %+v", fetched)
	}

	updateBody := map[string]any{
		"name":   "Alice Updated",
		"email":  "alice.updated@example.com",
		"age":    29,
		"status": "inactive",
	}
	update := doRequest(t, httpApp, http.MethodPut, "/api/v1/user/"+toString(created.ID), updateBody)
	var updated userResponse
	mustDecode(t, update.Data, &updated)
	if updated.Name != "Alice Updated" || updated.Status != "inactive" {
		t.Fatalf("unexpected updated user: %+v", updated)
	}

	deleteResponse := doRequest(t, httpApp, http.MethodDelete, "/api/v1/user/"+toString(created.ID), nil)
	if deleteResponse.Code != common.RETURN_SUCCESS {
		t.Fatalf("delete returned unexpected code: %+v", deleteResponse)
	}

	notFound := doRequest(t, httpApp, http.MethodGet, "/api/v1/user/"+toString(created.ID), nil)
	if notFound.Message != "user not found" {
		t.Fatalf("expected deleted user to be missing, got %+v", notFound)
	}
}

func writeIntegrationConfig(t *testing.T) (string, string) {
	t.Helper()

	tempDir := t.TempDir()
	databasePath := filepath.ToSlash(filepath.Join(tempDir, "demo.db"))
	configPath := filepath.Join(tempDir, "server.yaml")
	configBody := []byte("server:\n" +
		"  app_id: \"awesome-fiber-template\"\n" +
		"  app_name: \"Awesome-Fiber-Template\"\n" +
		"  app_version: \"v1.0.0\"\n" +
		"  mode: \"debug\"\n" +
		"  ip_address: \"127.0.0.1\"\n" +
		"  port: \"9999\"\n" +
		"  memory_limit: 1\n" +
		"  gc_percent: 1000\n" +
		"  network: \"tcp\"\n" +
		"  enable_prefork: false\n" +
		"  is_full_stack: false\n" +
		"database:\n" +
		"  enabled: true\n" +
		"  db_type: \"sqlite\"\n" +
		"  sqlite:\n" +
		"    path: \"" + databasePath + "\"\n" +
		"redis:\n" +
		"  enabled: false\n" +
		"prometheus:\n" +
		"  enabled: false\n" +
		"schedule:\n" +
		"  enabled: false\n" +
		"waf:\n" +
		"  enabled: false\n" +
		"middleware:\n" +
		"  cors:\n" +
		"    allow_origins: [\"http://127.0.0.1:8888\"]\n" +
		"  limiter:\n" +
		"    enabled: false\n")

	if err := os.WriteFile(configPath, configBody, 0o644); err != nil {
		t.Fatalf("write test config failed: %v", err)
	}

	return configPath, filepath.FromSlash(databasePath)
}

func doRequest(t *testing.T, app *fiber.App, method, path string, body any) apiResponse {
	t.Helper()
	return doRequestWithHeaders(t, app, method, path, body, nil)
}

func doRequestWithHeaders(t *testing.T, app *fiber.App, method, path string, body any, headers map[string]string) apiResponse {
	t.Helper()
	resp := rawRequest(t, app, method, path, body, headers)
	defer resp.Body.Close()

	var result apiResponse
	result.Status = resp.StatusCode
	result.Headers = resp.Header.Clone()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	return result
}

func rawRequest(t *testing.T, app *fiber.App, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body failed: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	return resp
}

func mustDecode(t *testing.T, raw json.RawMessage, target any) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}
}

func toString(id int64) string {
	return strconv.FormatInt(id, 10)
}
