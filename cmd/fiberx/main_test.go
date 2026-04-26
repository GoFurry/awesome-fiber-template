package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIOutputsV1SupportMatrix(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))

	output := captureStdout(t, func() error {
		return run([]string{"list", "presets"})
	})
	if !strings.Contains(output, "heavy\timplemented=true") {
		t.Fatalf("expected heavy preset to be listed as implemented, got:\n%s", output)
	}
	if !strings.Contains(output, "extra-light\timplemented=true") {
		t.Fatalf("expected extra-light preset to be listed as implemented, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"list", "capabilities"})
	})
	if !strings.Contains(output, "redis\timplemented=true") {
		t.Fatalf("expected redis capability to be listed as implemented, got:\n%s", output)
	}
	if !strings.Contains(output, "swagger\timplemented=true") {
		t.Fatalf("expected swagger capability to be listed as implemented, got:\n%s", output)
	}
	if !strings.Contains(output, "embedded-ui\timplemented=true") {
		t.Fatalf("expected embedded-ui capability to be listed as implemented, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"validate"})
	})
	if !strings.Contains(output, "state 1 generator validated successfully") {
		t.Fatalf("expected State 1 validate output, got:\n%s", output)
	}
	if !strings.Contains(output, "implemented presets: heavy,medium,light,extra-light") {
		t.Fatalf("expected implemented preset matrix, got:\n%s", output)
	}
	if !strings.Contains(output, "implemented capabilities: redis,swagger,embedded-ui") {
		t.Fatalf("expected implemented capability matrix, got:\n%s", output)
	}
	if !strings.Contains(output, "default medium experience: swagger,embedded-ui") {
		t.Fatalf("expected medium experience summary, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"doctor"})
	})
	if !strings.Contains(output, "state: state-1") || !strings.Contains(output, "phase: phase-6-medium-production-baseline") {
		t.Fatalf("expected State 1 / Phase 6 doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "medium-production-baseline: enabled") {
		t.Fatalf("expected medium production baseline flag in doctor output, got:\n%s", output)
	}
}

func TestCLIExplainAndGenerate(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))

	output := captureStdout(t, func() error {
		return run([]string{"explain", "preset", "heavy"})
	})
	if !strings.Contains(output, "implemented: true") || !strings.Contains(output, "packs: preset-heavy") {
		t.Fatalf("expected heavy explain output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"explain", "capability", "redis"})
	})
	if !strings.Contains(output, "implemented: true") || !strings.Contains(output, "allowed_presets: heavy,medium") {
		t.Fatalf("expected redis explain output, got:\n%s", output)
	}

	workdir := t.TempDir()
	withWorkingDir(t, workdir, func() {
		output = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "heavy"})
		})
	})
	if !strings.Contains(output, "generated preset=heavy") {
		t.Fatalf("expected heavy generation summary, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "main.go")); err != nil {
		t.Fatalf("expected heavy project to be generated: %v", err)
	}

	workdir = t.TempDir()
	withWorkingDir(t, workdir, func() {
		output = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "medium", "--with", "redis"})
		})
	})
	if !strings.Contains(output, "generated preset=medium") || !strings.Contains(output, "capabilities: swagger,embedded-ui,redis") {
		t.Fatalf("expected medium+redis generation summary, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "internal", "infra", "cache", "redis.go")); err != nil {
		t.Fatalf("expected redis capability file to be generated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "docs", "openapi.yaml")); err != nil {
		t.Fatalf("expected swagger asset to be generated for medium: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "internal", "transport", "http", "webui", "dist", "index.html")); err != nil {
		t.Fatalf("expected embedded-ui asset to be generated for medium: %v", err)
	}

	workdir = t.TempDir()
	withWorkingDir(t, workdir, func() {
		output = captureStdout(t, func() error {
			return run([]string{"init", "--preset", "extra-light"})
		})
	})
	if !strings.Contains(output, "generated preset=extra-light") {
		t.Fatalf("expected extra-light init summary, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(workdir, "config", "server.yaml")); err != nil {
		t.Fatalf("expected init to write into the current directory: %v", err)
	}
}

func manifestRootForCLI(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", "..", "generator"))
	if err != nil {
		t.Fatalf("resolve manifest root: %v", err)
	}
	return root
}

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	runErr := fn()
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, reader); err != nil {
		t.Fatalf("read stdout buffer: %v", err)
	}

	if runErr != nil {
		t.Fatalf("run() returned error: %v", runErr)
	}

	return buffer.String()
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current dir: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change dir to %q: %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore dir to %q: %v", original, err)
		}
	}()

	fn()
}
