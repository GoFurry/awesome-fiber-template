package core

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRunSupportsV1PresetMatrix(t *testing.T) {
	testCases := []struct {
		name         string
		preset       string
		capabilities []string
		routerPath   string
		routerSnippet string
		expectRedis  bool
		expectMedium bool
	}{
		{name: "heavy", preset: "heavy", routerPath: filepath.Join("internal", "http", "router.go"), routerSnippet: `return "heavy"`},
		{name: "heavy with redis", preset: "heavy", capabilities: []string{"redis"}, routerPath: filepath.Join("internal", "http", "router.go"), routerSnippet: `return "heavy"`, expectRedis: true},
		{name: "medium", preset: "medium", routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectMedium: true},
		{name: "medium with redis", preset: "medium", capabilities: []string{"redis"}, routerPath: filepath.Join("internal", "transport", "http", "router", "router.go"), routerSnippet: `registerSwaggerRoutes(app, deps.Config)`, expectRedis: true, expectMedium: true},
		{name: "light", preset: "light", routerPath: filepath.Join("internal", "http", "router.go"), routerSnippet: `return "light"`},
		{name: "extra-light", preset: "extra-light", routerPath: filepath.Join("internal", "http", "router.go"), routerSnippet: `return "extra-light"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			req := Request{
				ProjectName:  "demo",
				ModulePath:   "github.com/example/demo",
				Preset:       tc.preset,
				Capabilities: tc.capabilities,
				Options: map[string]string{
					"command":       "new",
					"manifest_root": "../../generator",
					"target_dir":    targetDir,
				},
			}

			summary, err := Run(req)
			if err != nil {
				t.Fatalf("Run() returned error: %v", err)
			}

			if summary.Preset != tc.preset {
				t.Fatalf("expected preset %q, got %q", tc.preset, summary.Preset)
			}
			if summary.TargetDir != targetDir {
				t.Fatalf("expected target dir %q, got %q", targetDir, summary.TargetDir)
			}

			assertGeneratedFileContains(t, targetDir, "README.md", tc.preset)
			assertGeneratedFileContains(t, targetDir, "README.md", "github.com/example/demo")
			assertGeneratedFileContains(t, targetDir, tc.routerPath, tc.routerSnippet)
			if tc.expectMedium {
				assertGeneratedFileContains(t, targetDir, filepath.Join("config", "server.yaml"), `route_prefix: "/docs"`)
				assertGeneratedFileContains(t, targetDir, filepath.Join("docs", "openapi.yaml"), "openapi: 3.0.3")
				assertGeneratedFileContains(t, targetDir, filepath.Join("internal", "transport", "http", "webui", "dist", "index.html"), "embedded UI ships")
			}

			bootstrap := readGeneratedFile(t, targetDir, filepath.Join("internal", "bootstrap", "bootstrap.go"))
			if tc.expectRedis {
				if !strings.Contains(bootstrap, `"cache:redis"`) {
					t.Fatalf("expected redis injection in bootstrap, got:\n%s", bootstrap)
				}
			} else if strings.Contains(bootstrap, `"cache:redis"`) {
				t.Fatalf("did not expect redis injection in bootstrap, got:\n%s", bootstrap)
			}
			if tc.expectMedium {
				if !strings.Contains(bootstrap, `"docs:swagger"`) || !strings.Contains(bootstrap, `"ui:embedded"`) {
					t.Fatalf("expected default medium capabilities in bootstrap, got:\n%s", bootstrap)
				}
			}

			runGeneratedProjectTests(t, targetDir)
			if tc.expectMedium {
				runMediumBlackBoxScenario(t, targetDir, tc.expectRedis)
			}
		})
	}
}

func TestGenerateRejectsUnsupportedCombinations(t *testing.T) {
	testCases := []struct {
		name         string
		preset       string
		capabilities []string
		want         string
	}{
		{name: "light with redis", preset: "light", capabilities: []string{"redis"}, want: `not allowed for preset "light"`},
		{name: "extra-light with redis", preset: "extra-light", capabilities: []string{"redis"}, want: `not allowed for preset "extra-light"`},
		{name: "light with embedded-ui", preset: "light", capabilities: []string{"embedded-ui"}, want: `not allowed for preset "light"`},
		{name: "heavy with swagger", preset: "heavy", capabilities: []string{"swagger"}, want: `not allowed for preset "heavy"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := Request{
				ProjectName:  "demo",
				ModulePath:   "github.com/example/demo",
				Preset:       tc.preset,
				Capabilities: tc.capabilities,
				Options: map[string]string{
					"manifest_root": "../../generator",
					"target_dir":    t.TempDir(),
				},
			}

			err := Generate(req)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func assertGeneratedFileContains(t *testing.T, targetDir string, relativePath string, want string) {
	t.Helper()

	content := readGeneratedFile(t, targetDir, relativePath)
	if !strings.Contains(content, want) {
		t.Fatalf("expected %s to contain %q, got:\n%s", relativePath, want, content)
	}
}

func readGeneratedFile(t *testing.T, targetDir string, relativePath string) string {
	t.Helper()

	path := filepath.Join(targetDir, relativePath)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated file %q: %v", path, err)
	}
	return string(data)
}

func runGeneratedProjectTests(t *testing.T, targetDir string) {
	t.Helper()

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = targetDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated project go test failed: %v\n%s", err, string(output))
	}
}

func runMediumBlackBoxScenario(t *testing.T, targetDir string, enableRedis bool) {
	t.Helper()

	port := randomPort(t)
	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "data", "app.db")
	logPath := filepath.Join(tempDir, "logs", "app.log")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		t.Fatalf("create database dir failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("create log dir failed: %v", err)
	}

	configBody := `server:
  app_id: "fiberx"
  app_name: "demo"
  app_version: "v1.0.0"
  mode: "debug"
  ip_address: "127.0.0.1"
  port: "` + port + `"
database:
  enabled: true
  auto_migrate: true
  db_type: "sqlite"
  sqlite:
    path: "` + filepath.ToSlash(databasePath) + `"
log:
  log_level: "debug"
  log_mode: "text"
  log_path: "` + filepath.ToSlash(logPath) + `"
middleware:
  cors:
    allow_origins:
      - "*"
  gzip:
    enabled: true
swagger:
  enabled: true
  route_prefix: "/docs"
embedded_ui:
  enabled: true
  route_prefix: "/ui"
redis:
  enabled: ` + strconv.FormatBool(enableRedis) + `
  addr: "127.0.0.1:0"
  password: ""
  db: 0
`
	configPath := filepath.Join(tempDir, "server.yaml")
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	binaryPath := buildBinary(t, targetDir)
	cmd := exec.Command(binaryPath, "serve", "--config", configPath)
	cmd.Dir = targetDir
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start medium service failed: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	baseURL := "http://127.0.0.1:" + port
	waitForReady(t, baseURL+"/healthz", &output, cmd)

	health := doJSONRequest(t, "GET", baseURL+"/healthz", nil, nil)
	if health.StatusCode != 200 || health.Code != 1 {
		t.Fatalf("unexpected health response: %+v", health)
	}
	if health.Headers.Get("X-Request-ID") == "" || health.Headers.Get("X-Content-Type-Options") == "" {
		t.Fatalf("expected request-id and security headers on health response")
	}
	if enableRedis && !healthBodyContainsService(t, health.Data, "cache:redis") {
		t.Fatalf("expected redis service to appear in health payload: %s", string(health.Data))
	}

	for _, path := range []string{"/livez", "/readyz", "/startupz"} {
		resp := doJSONRequest(t, "GET", baseURL+path, nil, nil)
		if resp.StatusCode != 200 {
			t.Fatalf("expected %s to return 200, got %+v", path, resp)
		}
	}

	docsResp, err := http.Get(baseURL + "/docs/openapi.yaml")
	if err != nil {
		t.Fatalf("fetch docs failed: %v", err)
	}
	body, _ := io.ReadAll(docsResp.Body)
	_ = docsResp.Body.Close()
	if docsResp.StatusCode != 200 || !strings.Contains(string(body), "openapi: 3.0.3") {
		t.Fatalf("unexpected docs response: status=%d body=%s", docsResp.StatusCode, string(body))
	}

	uiResp, err := http.Get(baseURL + "/ui")
	if err != nil {
		t.Fatalf("fetch ui failed: %v", err)
	}
	uiBody, _ := io.ReadAll(uiResp.Body)
	_ = uiResp.Body.Close()
	if uiResp.StatusCode != 200 || !strings.Contains(string(uiBody), "embedded UI ships") {
		t.Fatalf("unexpected ui response: status=%d body=%s", uiResp.StatusCode, string(uiBody))
	}

	createPayload := map[string]any{
		"name":   "Alice",
		"email":  "alice@example.com",
		"age":    28,
		"status": "active",
	}
	create := doJSONRequest(t, "POST", baseURL+"/api/v1/user/", createPayload, nil)
	if create.StatusCode != 200 || create.Code != 1 {
		t.Fatalf("create user failed: %+v", create)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(create.Data, &created); err != nil {
		t.Fatalf("decode created user failed: %v", err)
	}

	for index := 0; index < 8; index++ {
		payload := map[string]any{
			"name":   "User " + strconv.Itoa(index),
			"email":  "user" + strconv.Itoa(index) + "@example.com",
			"age":    20 + index,
			"status": "active",
		}
		resp := doJSONRequest(t, "POST", baseURL+"/api/v1/user/", payload, nil)
		if resp.StatusCode != 200 || resp.Code != 1 {
			t.Fatalf("seed user %d failed: %+v", index, resp)
		}
	}

	list := doJSONRequest(t, "GET", baseURL+"/api/v1/user/?page_num=1&page_size=20", nil, nil)
	if list.StatusCode != 200 || list.Code != 1 || list.Headers.Get("ETag") == "" {
		t.Fatalf("list users failed: %+v", list)
	}

	compressed, err := rawRequest("GET", baseURL+"/api/v1/user/?page_num=1&page_size=20", nil, map[string]string{"Accept-Encoding": "gzip"})
	if err != nil {
		t.Fatalf("compressed request failed: %v", err)
	}
	_, _ = io.ReadAll(compressed.Body)
	_ = compressed.Body.Close()
	if compressed.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip response, got headers %#v", compressed.Header)
	}

	get := doJSONRequest(t, "GET", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), nil, nil)
	if get.StatusCode != 200 || get.Code != 1 {
		t.Fatalf("get user failed: %+v", get)
	}

	update := doJSONRequest(t, "PUT", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), map[string]any{
		"name":   "Alice Updated",
		"email":  "alice.updated@example.com",
		"age":    29,
		"status": "inactive",
	}, nil)
	if update.StatusCode != 200 || update.Code != 1 {
		t.Fatalf("update user failed: %+v", update)
	}

	deleteResponse := doJSONRequest(t, "DELETE", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), nil, nil)
	if deleteResponse.StatusCode != 200 || deleteResponse.Code != 1 {
		t.Fatalf("delete user failed: %+v", deleteResponse)
	}

	notFound := doJSONRequest(t, "GET", baseURL+"/api/v1/user/"+strconv.FormatInt(created.ID, 10), nil, nil)
	if notFound.StatusCode != 404 || notFound.Code == 1 || !strings.Contains(strings.ToLower(notFound.Message), "not found") {
		t.Fatalf("expected deleted user to be missing, got %+v", notFound)
	}

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("expected sqlite database at %s: %v", databasePath, err)
	}
}

type apiResponse struct {
	StatusCode int
	Headers    http.Header
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
}

func doJSONRequest(t *testing.T, method string, url string, body any, headers map[string]string) apiResponse {
	t.Helper()

	resp, err := rawRequest(method, url, body, headers)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()

	var decoded apiResponse
	decoded.StatusCode = resp.StatusCode
	decoded.Headers = resp.Header.Clone()
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	return decoded
}

func rawRequest(method string, url string, body any, headers map[string]string) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(raw))
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return (&http.Client{Timeout: 10 * time.Second}).Do(req)
}

func buildBinary(t *testing.T, targetDir string) string {
	t.Helper()

	binaryName := "service"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(t.TempDir(), binaryName)
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = targetDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("build binary failed: %v", err)
	}
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("expected built binary to exist: %v", err)
	}
	return binaryPath
}

func waitForReady(t *testing.T, url string, output *strings.Builder, cmd *exec.Cmd) {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			t.Fatalf("service exited before readiness:\n%s", output.String())
		}

		resp, err := http.Get(url)
		if err == nil {
			_, _ = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	t.Fatalf("service did not become ready:\n%s", output.String())
}

func healthBodyContainsService(t *testing.T, raw json.RawMessage, want string) bool {
	t.Helper()

	var payload struct {
		Services []string `json:"services"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode health payload failed: %v", err)
	}
	for _, service := range payload.Services {
		if service == want {
			return true
		}
	}
	return false
}

func randomPort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port failed: %v", err)
	}
	defer listener.Close()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("parse port failed: %v", err)
	}
	return port
}
