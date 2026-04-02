package bootstrap_test

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	env "github.com/GoFurry/awesome-fiber-template/v3/extra-light/config"
	"github.com/GoFurry/awesome-fiber-template/v3/extra-light/internal/bootstrap"
	apphttp "github.com/GoFurry/awesome-fiber-template/v3/extra-light/internal/http"
	"github.com/gofiber/fiber/v3"
)

func TestBootstrapAndHealthEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "server.yaml")
	dbPath := filepath.Join(tempDir, "data", "app.db")
	logPath := filepath.Join(tempDir, "logs", "app.log")

	configContent := []byte("server:\n  app_name: extra-light-test\n  mode: debug\n  ip_address: 127.0.0.1\n  port: 9999\ndatabase:\n  path: " + quoteYAML(dbPath) + "\nlog:\n  log_level: debug\n  log_path: " + quoteYAML(logPath) + "\n")
	if err := os.WriteFile(configPath, configContent, 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	env.MustInitServerConfig(configPath)
	if err := bootstrap.Start(); err != nil {
		t.Fatalf("bootstrap start failed: %v", err)
	}
	t.Cleanup(func() {
		if err := bootstrap.Shutdown(); err != nil {
			t.Fatalf("bootstrap shutdown failed: %v", err)
		}
	})

	app := apphttp.New()

	assertStatus(t, app, "/livez", fiber.StatusOK)
	assertStatus(t, app, "/readyz", fiber.StatusOK)
	assertStatus(t, app, "/startupz", fiber.StatusOK)
	assertStatus(t, app, "/healthz", fiber.StatusOK)

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected sqlite file to exist: %v", err)
	}
}

func assertStatus(t *testing.T, app *fiber.App, path string, want int) {
	t.Helper()

	request := httptest.NewRequest(fiber.MethodGet, path, nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request %s failed: %v", path, err)
	}
	defer response.Body.Close()

	if response.StatusCode != want {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("unexpected status for %s: got %d want %d body=%s", path, response.StatusCode, want, string(body))
	}
}

func quoteYAML(value string) string {
	return "\"" + filepath.ToSlash(value) + "\""
}
