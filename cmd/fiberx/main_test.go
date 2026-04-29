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
	if !strings.Contains(output, "state 4 generator validated successfully") {
		t.Fatalf("expected State 4 validate output, got:\n%s", output)
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
	if !strings.Contains(output, "stable production baseline: medium") || !strings.Contains(output, "completed production track: heavy") {
		t.Fatalf("expected production track summary, got:\n%s", output)
	}
	if !strings.Contains(output, "current stage: phase-13-version-upgrade-and-diff-detection") || !strings.Contains(output, "phase 10 delivery: completed") || !strings.Contains(output, "phase 11 delivery: completed") || !strings.Contains(output, "phase 12 delivery: completed") || !strings.Contains(output, "phase 13 focus: generator/template versioning and diff detection") {
		t.Fatalf("expected phase 13 summary with completed phase 12 delivery, got:\n%s", output)
	}
	if !strings.Contains(output, "default heavy experience: swagger,embedded-ui") {
		t.Fatalf("expected heavy experience summary, got:\n%s", output)
	}
	if !strings.Contains(output, "light optional experience: swagger,embedded-ui") || !strings.Contains(output, "extra-light optional experience: none") {
		t.Fatalf("expected light/extra-light capability summary, got:\n%s", output)
	}
	if !strings.Contains(output, "capability-policy-swagger: default=heavy,medium optional=light unsupported=extra-light") ||
		!strings.Contains(output, "capability-policy-embedded-ui: default=heavy,medium optional=light unsupported=extra-light") ||
		!strings.Contains(output, "capability-policy-redis: default=(none) optional=heavy,medium unsupported=light,extra-light") {
		t.Fatalf("expected capability policy summary, got:\n%s", output)
	}
	if !strings.Contains(output, "default stack: fiber-v3 + cobra + viper") || !strings.Contains(output, "supported fiber versions: v3,v2") || !strings.Contains(output, "supported cli styles: cobra,native") || !strings.Contains(output, "default logger: zap") || !strings.Contains(output, "default database: sqlite") || !strings.Contains(output, "default data access: stdlib") {
		t.Fatalf("expected stack policy summary, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"doctor"})
	})
	if !strings.Contains(output, "state: state-4") || !strings.Contains(output, "phase: phase-13-version-upgrade-and-diff-detection") {
		t.Fatalf("expected State 4 / Phase 13 doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "medium-production-baseline: stable") || !strings.Contains(output, "heavy-production-track: completed") {
		t.Fatalf("expected medium production baseline flag in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "phase-9-stack-normalization: completed") || !strings.Contains(output, "phase-10-capability-consolidation: completed") || !strings.Contains(output, "phase-11-runtime-options-and-data-access: completed") || !strings.Contains(output, "phase-12-capability-level-verification: completed") || !strings.Contains(output, "phase-13-version-upgrade-and-diff-detection: active") || !strings.Contains(output, "phase-13-focus: generator-template-versioning-and-diff-detection") {
		t.Fatalf("expected phase 12 completed and phase 13 active flags in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "default-heavy-capabilities: swagger,embedded-ui") {
		t.Fatalf("expected heavy defaults in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "light-optional-capabilities: swagger,embedded-ui") || !strings.Contains(output, "extra-light-optional-capabilities: none") {
		t.Fatalf("expected light/extra-light defaults in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "capability-policy-swagger: default=heavy,medium optional=light unsupported=extra-light") ||
		!strings.Contains(output, "capability-policy-embedded-ui: default=heavy,medium optional=light unsupported=extra-light") ||
		!strings.Contains(output, "capability-policy-redis: default=(none) optional=heavy,medium unsupported=light,extra-light") {
		t.Fatalf("expected capability policy in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "default-stack: fiber-v3 + cobra + viper") || !strings.Contains(output, "supported-fiber-versions: v3,v2") || !strings.Contains(output, "supported-cli-styles: cobra,native") || !strings.Contains(output, "default-logger: zap") || !strings.Contains(output, "default-database: sqlite") || !strings.Contains(output, "default-data-access: stdlib") {
		t.Fatalf("expected stack policy in doctor output, got:\n%s", output)
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
	if !strings.Contains(output, "default_capabilities: swagger,embedded-ui") || !strings.Contains(output, "allowed_capabilities: swagger,embedded-ui,redis") {
		t.Fatalf("expected heavy default capability output, got:\n%s", output)
	}
	if !strings.Contains(output, "default_stack: fiber-v3 + cobra + viper") || !strings.Contains(output, "default_logger: zap") || !strings.Contains(output, "default_database: sqlite") || !strings.Contains(output, "default_data_access: stdlib") || !strings.Contains(output, "supported_loggers: zap,slog") || !strings.Contains(output, "supported_databases: sqlite,pgsql,mysql") || !strings.Contains(output, "supported_data_access: stdlib,sqlx,sqlc") {
		t.Fatalf("expected heavy stack explain output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"explain", "preset", "light"})
	})
	if !strings.Contains(output, "implemented: true") || !strings.Contains(output, "packs: preset-light") {
		t.Fatalf("expected light explain output, got:\n%s", output)
	}
	if !strings.Contains(output, "default_capabilities: (none)") || !strings.Contains(output, "allowed_capabilities: swagger,embedded-ui") || !strings.Contains(output, "supported_loggers: zap,slog") {
		t.Fatalf("expected light capability output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"explain", "preset", "extra-light"})
	})
	if !strings.Contains(output, "implemented: true") || !strings.Contains(output, "packs: preset-extra-light") {
		t.Fatalf("expected extra-light explain output, got:\n%s", output)
	}
	if !strings.Contains(output, "default_capabilities: (none)") || !strings.Contains(output, "allowed_capabilities: (none)") || !strings.Contains(output, "phase11_runtime_options: unsupported") {
		t.Fatalf("expected extra-light capability output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"explain", "capability", "swagger"})
	})
	if !strings.Contains(output, "allowed_presets: heavy,medium,light") ||
		!strings.Contains(output, "default_on_presets: heavy,medium") ||
		!strings.Contains(output, "optional_on_presets: light") ||
		!strings.Contains(output, "unsupported_on_presets: extra-light") {
		t.Fatalf("expected swagger explain boundary output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"explain", "capability", "embedded-ui"})
	})
	if !strings.Contains(output, "allowed_presets: heavy,medium,light") ||
		!strings.Contains(output, "default_on_presets: heavy,medium") ||
		!strings.Contains(output, "optional_on_presets: light") ||
		!strings.Contains(output, "unsupported_on_presets: extra-light") ||
		!strings.Contains(output, "depends_on: (none)") {
		t.Fatalf("expected embedded-ui explain boundary output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"explain", "capability", "redis"})
	})
	if !strings.Contains(output, "implemented: true") ||
		!strings.Contains(output, "allowed_presets: heavy,medium") ||
		!strings.Contains(output, "default_on_presets: (none)") ||
		!strings.Contains(output, "optional_on_presets: heavy,medium") ||
		!strings.Contains(output, "unsupported_on_presets: light,extra-light") {
		t.Fatalf("expected redis explain output, got:\n%s", output)
	}

	workdir := t.TempDir()
	withWorkingDir(t, workdir, func() {
		output = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "heavy"})
		})
	})
	if !strings.Contains(output, "generated preset=heavy") || !strings.Contains(output, "capabilities: swagger,embedded-ui") || !strings.Contains(output, "stack: fiber-v3 + cobra + viper") || !strings.Contains(output, "runtime: logger=zap db=sqlite data-access=stdlib") {
		t.Fatalf("expected heavy generation summary, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "main.go")); err != nil {
		t.Fatalf("expected heavy project to be generated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "cmd", "root.go")); err != nil {
		t.Fatalf("expected cobra command root to be generated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "docs", "runbook.md")); err != nil {
		t.Fatalf("expected generated runbook to be generated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "config", "server.prod.yaml")); err != nil {
		t.Fatalf("expected generated prod config profile to be generated: %v", err)
	}

	workdir = t.TempDir()
	withWorkingDir(t, workdir, func() {
		output = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "medium", "--with", "redis"})
		})
	})
	if !strings.Contains(output, "generated preset=medium") || !strings.Contains(output, "capabilities: swagger,embedded-ui,redis") || !strings.Contains(output, "runtime overlays: runtime-logger-zap,runtime-data-stdlib") {
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
			return run([]string{"new", "demo", "--preset", "medium", "--fiber-version", "v2", "--cli-style", "native"})
		})
	})
	if !strings.Contains(output, "stack: fiber-v2 + native") || !strings.Contains(output, "runtime: logger=zap db=sqlite data-access=stdlib") {
		t.Fatalf("expected compatibility stack generation summary, got:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(workdir, "demo", "main.go")); err != nil {
		t.Fatalf("expected compatibility project to be generated: %v", err)
	}

	workdir = t.TempDir()
	withWorkingDir(t, workdir, func() {
		output = captureStdout(t, func() error {
			return run([]string{"init", "--preset", "extra-light", "--fiber-version", "v2", "--cli-style", "native"})
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
