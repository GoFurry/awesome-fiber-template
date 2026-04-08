package internaltest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type APIResponse struct {
	Status  int             `json:"-"`
	Headers http.Header     `json:"-"`
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type UserResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Status string `json:"status"`
}

type UserListResponse struct {
	Total int64          `json:"total"`
	List  []UserResponse `json:"list"`
}

func TemplateRoot(t *testing.T, template string) string {
	t.Helper()

	root, err := workspaceRoot()
	if err != nil {
		t.Fatalf("resolve workspace root failed: %v", err)
	}
	return filepath.Join(root, template)
}

func workspaceRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve caller path failed")
	}
	// helpers.go -> internaltest -> test -> v3
	return filepath.Dir(filepath.Dir(filepath.Dir(file))), nil
}

func StartService(t *testing.T, workdir, configPath, baseURL, readinessPath string) func() {
	t.Helper()

	binaryPath := buildBinary(t, workdir)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binaryPath, "serve", "--config", configPath)
	cmd.Dir = workdir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start service failed: %v", err)
	}

	var (
		exitMu  sync.Mutex
		exitErr error
		done    = make(chan struct{})
	)
	go func() {
		defer close(done)
		exitMu.Lock()
		exitErr = cmd.Wait()
		exitMu.Unlock()
	}()

	waitForHTTP(t, baseURL+readinessPath, done, func() error {
		exitMu.Lock()
		defer exitMu.Unlock()
		return exitErr
	}, stdout.String, stderr.String)

	return func() {
		cancel()
		select {
		case <-done:
			exitMu.Lock()
			err := exitErr
			exitMu.Unlock()
			if err != nil && !isExpectedExit(err) {
				t.Fatalf("service stopped unexpectedly: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
			}
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			select {
			case <-done:
				exitMu.Lock()
				err := exitErr
				exitMu.Unlock()
				if err != nil && !isExpectedExit(err) {
					t.Fatalf("service stopped unexpectedly: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("service did not stop in time\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
			}
		}
	}
}

func buildBinary(t *testing.T, workdir string) string {
	t.Helper()

	binDir := t.TempDir()
	binaryName := "service"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(binDir, binaryName)

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = workdir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("build service binary failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	return binaryPath
}

func waitForHTTP(t *testing.T, url string, done <-chan struct{}, exitErr func() error, stdout, stderr func() string) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		select {
		case <-done:
			if err := exitErr(); err != nil && !isExpectedExit(err) {
				t.Fatalf("service exited before becoming ready at %s: %v\nstdout:\n%s\nstderr:\n%s", url, err, stdout(), stderr())
			}
			t.Fatalf("service exited before becoming ready at %s\nstdout:\n%s\nstderr:\n%s", url, stdout(), stderr())
		default:
		}

		resp, err := client.Get(url)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	t.Fatalf("service did not become ready at %s\nstdout:\n%s\nstderr:\n%s", url, stdout(), stderr())
}

func DoRequest(t *testing.T, method, url string, body any) APIResponse {
	t.Helper()
	return DoRequestWithHeaders(t, method, url, body, nil)
}

func DoRequestWithHeaders(t *testing.T, method, url string, body any, headers map[string]string) APIResponse {
	t.Helper()
	resp := RawRequest(t, method, url, body, headers)
	defer resp.Body.Close()

	var result APIResponse
	result.Status = resp.StatusCode
	result.Headers = resp.Header.Clone()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	return result
}

func RawRequest(t *testing.T, method, url string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var reader io.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body failed: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("build request failed: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	return resp
}

func MustDecode(t *testing.T, raw json.RawMessage, target any) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}
}

func ToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

func AssertStatus(t *testing.T, url string, want int) {
	t.Helper()

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(url)
	if err != nil {
		t.Fatalf("request %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status for %s: got %d want %d body=%s", url, resp.StatusCode, want, string(body))
	}
}

func WaitForStatus(t *testing.T, url string, want int) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := (&http.Client{Timeout: 2 * time.Second}).Get(url)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == want {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	t.Fatalf("status %d not observed at %s within timeout", want, url)
}

func FormatBaseURL(port string) string {
	return fmt.Sprintf("http://127.0.0.1:%s", port)
}

func isExpectedExit(err error) bool {
	if err == nil {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "signal: killed") ||
		strings.Contains(lower, "terminated") ||
		strings.Contains(lower, "exit status")
}
