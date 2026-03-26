package bootstrap

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/config"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/infra/db"
	usermodels "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/models"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/transport/http/router"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type apiResponse struct {
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

	runtimeApp, err := Start()
	if err != nil {
		t.Fatalf("start bootstrap failed: %v", err)
	}

	httpApp := router.New().Init(runtimeApp.RouteModules...)
	t.Cleanup(func() {
		_ = httpApp.Shutdown()
		_ = Shutdown()
	})

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("sqlite database file not created: %v", err)
	}
	if !db.Orm.DB().Migrator().HasTable(&usermodels.User{}) {
		t.Fatalf("expected demo_users table to be auto migrated")
	}
	if !db.Orm.DB().Migrator().HasTable("schema_migrations") {
		t.Fatalf("expected schema_migrations table to be created")
	}

	health := doRequest(t, httpApp, http.MethodGet, "/healthz", nil)
	if health.Code != common.RETURN_SUCCESS {
		t.Fatalf("healthz returned unexpected code: %+v", health)
	}

	createBody := map[string]any{
		"name":   "Alice",
		"email":  "alice@example.com",
		"age":    28,
		"status": "active",
	}
	create := doRequest(t, httpApp, http.MethodPost, "/api/v1/users/", createBody)
	var created userResponse
	mustDecode(t, create.Data, &created)
	if created.ID == 0 {
		t.Fatalf("expected created user id, got %+v", created)
	}

	list := doRequest(t, httpApp, http.MethodGet, "/api/v1/users/?page_num=1&page_size=10", nil)
	var users userListResponse
	mustDecode(t, list.Data, &users)
	if users.Total < 2 || len(users.List) < 2 {
		t.Fatalf("unexpected list payload: %+v", users)
	}

	get := doRequest(t, httpApp, http.MethodGet, "/api/v1/users/"+toString(created.ID), nil)
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
	update := doRequest(t, httpApp, http.MethodPut, "/api/v1/users/"+toString(created.ID), updateBody)
	var updated userResponse
	mustDecode(t, update.Data, &updated)
	if updated.Name != "Alice Updated" || updated.Status != "inactive" {
		t.Fatalf("unexpected updated user: %+v", updated)
	}

	deleteResponse := doRequest(t, httpApp, http.MethodDelete, "/api/v1/users/"+toString(created.ID), nil)
	if deleteResponse.Code != common.RETURN_SUCCESS {
		t.Fatalf("delete returned unexpected code: %+v", deleteResponse)
	}

	notFound := doRequest(t, httpApp, http.MethodGet, "/api/v1/users/"+toString(created.ID), nil)
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

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	return result
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
