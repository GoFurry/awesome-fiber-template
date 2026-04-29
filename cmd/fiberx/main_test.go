package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoFurry/fiberx/internal/version"
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
	if !strings.Contains(output, "current stage: phase-15-build-and-post-generation-engineering") || !strings.Contains(output, "phase 10 delivery: completed") || !strings.Contains(output, "phase 11 delivery: completed") || !strings.Contains(output, "phase 12 delivery: completed") || !strings.Contains(output, "phase 13 delivery: completed") || !strings.Contains(output, "phase 14 delivery: completed") || !strings.Contains(output, "phase 15 focus: build and post-generation engineering") {
		t.Fatalf("expected phase 15 summary with completed phase 14 delivery, got:\n%s", output)
	}
	if !strings.Contains(output, "phase 15 delivery target: fiberx build and release-oriented output management") {
		t.Fatalf("expected phase 15 delivery target output, got:\n%s", output)
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
	if !strings.Contains(output, "state: state-4") || !strings.Contains(output, "phase: phase-15-build-and-post-generation-engineering") {
		t.Fatalf("expected State 4 / Phase 15 doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "medium-production-baseline: stable") || !strings.Contains(output, "heavy-production-track: completed") {
		t.Fatalf("expected medium production baseline flag in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "phase-9-stack-normalization: completed") || !strings.Contains(output, "phase-10-capability-consolidation: completed") || !strings.Contains(output, "phase-11-runtime-options-and-data-access: completed") || !strings.Contains(output, "phase-12-capability-level-verification: completed") || !strings.Contains(output, "phase-13-version-upgrade-and-diff-detection: completed") || !strings.Contains(output, "phase-14-upgrade-assistant-and-compatibility-policy: completed") || !strings.Contains(output, "phase-15-build-and-post-generation-engineering: active") || !strings.Contains(output, "phase-15-focus: build-and-post-generation-engineering") || !strings.Contains(output, "phase-15-delivery-target: fiberx-build-and-release-oriented-output-management") {
		t.Fatalf("expected phase 14 completed and phase 15 active flags in doctor output, got:\n%s", output)
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
	if !strings.Contains(output, "generated preset=heavy") || !strings.Contains(output, "capabilities: swagger,embedded-ui") || !strings.Contains(output, "stack: fiber-v3 + cobra + viper") || !strings.Contains(output, "runtime: logger=zap db=sqlite data-access=stdlib") || !strings.Contains(output, "metadata: .fiberx/manifest.json") {
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

func TestCLIInspectAndDiff(t *testing.T) {
	originalRoot := manifestRootForCLI(t)
	t.Setenv("FIBERX_MANIFEST_ROOT", originalRoot)

	workdir := t.TempDir()
	var generationOutput string
	withWorkingDir(t, workdir, func() {
		generationOutput = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "light"})
		})
	})
	if !strings.Contains(generationOutput, "metadata: .fiberx/manifest.json") {
		t.Fatalf("expected metadata path in generation summary, got:\n%s", generationOutput)
	}

	projectDir := filepath.Join(workdir, "demo")
	if _, err := os.Stat(filepath.Join(projectDir, ".fiberx", "manifest.json")); err != nil {
		t.Fatalf("expected generated metadata manifest: %v", err)
	}

	output := captureStdout(t, func() error {
		return run([]string{"inspect", projectDir})
	})
	if !strings.Contains(output, "generator-version:") || !strings.Contains(output, "template fingerprint:") || !strings.Contains(output, "managed files:") {
		t.Fatalf("expected inspect text output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"inspect", projectDir, "--json"})
	})
	var inspectPayload map[string]any
	if err := json.Unmarshal([]byte(output), &inspectPayload); err != nil {
		t.Fatalf("unmarshal inspect json: %v\n%s", err, output)
	}
	if inspectPayload["schema_version"] != "v1" {
		t.Fatalf("expected schema_version v1, got %#v", inspectPayload["schema_version"])
	}
	recipePayload, ok := inspectPayload["recipe"].(map[string]any)
	if !ok || recipePayload["preset"] != "light" {
		t.Fatalf("expected inspect recipe preset=light, got %#v", inspectPayload["recipe"])
	}

	output = captureStdout(t, func() error {
		return run([]string{"diff", projectDir})
	})
	if !strings.Contains(output, "status: clean") {
		t.Fatalf("expected clean diff output, got:\n%s", output)
	}

	output = captureStdout(t, func() error {
		return run([]string{"diff", projectDir, "--json"})
	})
	var diffPayload map[string]any
	if err := json.Unmarshal([]byte(output), &diffPayload); err != nil {
		t.Fatalf("unmarshal diff json: %v\n%s", err, output)
	}
	if diffPayload["status"] != "clean" {
		t.Fatalf("expected clean diff json status, got %#v", diffPayload["status"])
	}

	readmePath := filepath.Join(projectDir, "README.md")
	readmeData, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read generated README: %v", err)
	}
	if err := os.WriteFile(readmePath, append(readmeData, []byte("\nlocal drift\n")...), 0o644); err != nil {
		t.Fatalf("write local drift README: %v", err)
	}

	output = captureStdout(t, func() error {
		return run([]string{"diff", projectDir})
	})
	if !strings.Contains(output, "status: local_modified") || !strings.Contains(output, "changed files: README.md") {
		t.Fatalf("expected local_modified diff output, got:\n%s", output)
	}

	generatorCopy := filepath.Join(t.TempDir(), "generator")
	copyDir(t, originalRoot, generatorCopy)
	baseReadmePath := filepath.Join(generatorCopy, "assets", "base", "service-base-cobra", "README.md.tmpl")
	baseReadmeData, err := os.ReadFile(baseReadmePath)
	if err != nil {
		t.Fatalf("read copied generator README template: %v", err)
	}
	updatedReadme := strings.Replace(string(baseReadmeData), "Generated by `fiberx`.", "Generated by `fiberx` (phase13 drift).", 1)
	if err := os.WriteFile(baseReadmePath, []byte(updatedReadme), 0o644); err != nil {
		t.Fatalf("write copied generator README template: %v", err)
	}

	t.Setenv("FIBERX_MANIFEST_ROOT", generatorCopy)

	driftWorkdir := t.TempDir()
	withWorkingDir(t, driftWorkdir, func() {
		_ = captureStdout(t, func() error {
			return run([]string{"new", "demo", "--preset", "light"})
		})
	})
	driftProjectDir := filepath.Join(driftWorkdir, "demo")

	t.Setenv("FIBERX_MANIFEST_ROOT", generatorCopy)
	output = captureStdout(t, func() error {
		return run([]string{"diff", driftProjectDir})
	})
	if !strings.Contains(output, "status: clean") {
		t.Fatalf("expected clean diff against copied generator, got:\n%s", output)
	}

	t.Setenv("FIBERX_MANIFEST_ROOT", originalRoot)
	output = captureStdout(t, func() error {
		return run([]string{"diff", driftProjectDir})
	})
	if !strings.Contains(output, "status: generator_drift") || !strings.Contains(output, "generator drift files: README.md") {
		t.Fatalf("expected generator_drift output, got:\n%s", output)
	}

	if err := os.WriteFile(filepath.Join(driftProjectDir, "README.md"), append([]byte(updatedReadme), []byte("\nlocal drift too\n")...), 0o644); err != nil {
		t.Fatalf("write local+generator drift README: %v", err)
	}
	output = captureStdout(t, func() error {
		return run([]string{"diff", driftProjectDir})
	})
	if !strings.Contains(output, "status: local_and_generator_drift") || !strings.Contains(output, "changed files: README.md") || !strings.Contains(output, "generator drift files: README.md") {
		t.Fatalf("expected local_and_generator_drift output, got:\n%s", output)
	}
}

func TestCLIInspectAndDiffRejectMissingMetadata(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))
	workdir := t.TempDir()

	err := run([]string{"inspect", workdir})
	if err == nil || !strings.Contains(err.Error(), ".fiberx/manifest.json") {
		t.Fatalf("expected inspect missing metadata error, got %v", err)
	}

	err = run([]string{"diff", workdir})
	if err == nil || !strings.Contains(err.Error(), ".fiberx/manifest.json") {
		t.Fatalf("expected diff missing metadata error, got %v", err)
	}
}

func TestCLIUpgradeInspectAndPlan(t *testing.T) {
	originalRoot := manifestRootForCLI(t)
	t.Setenv("FIBERX_MANIFEST_ROOT", originalRoot)

	withGeneratorIdentityForCLI(t, "v0.13.0", "phase13-generated", func() {
		workdir := t.TempDir()
		withWorkingDir(t, workdir, func() {
			_ = captureStdout(t, func() error {
				return run([]string{"new", "demo", "--preset", "light"})
			})
		})
		projectDir := filepath.Join(workdir, "demo")

		withGeneratorIdentityForCLI(t, "v0.14.0", "phase14-current", func() {
			output := captureStdout(t, func() error {
				return run([]string{"upgrade", "inspect", projectDir})
			})
			if !strings.Contains(output, "compatibility level: compatible") || !strings.Contains(output, "diff status: clean") {
				t.Fatalf("expected compatible clean upgrade inspect output, got:\n%s", output)
			}

			output = captureStdout(t, func() error {
				return run([]string{"upgrade", "inspect", projectDir, "--json"})
			})
			var inspectPayload map[string]any
			if err := json.Unmarshal([]byte(output), &inspectPayload); err != nil {
				t.Fatalf("unmarshal upgrade inspect json: %v\n%s", err, output)
			}
			if inspectPayload["compatibility_level"] != "compatible" {
				t.Fatalf("expected compatible level in inspect json, got %#v", inspectPayload["compatibility_level"])
			}

			output = captureStdout(t, func() error {
				return run([]string{"upgrade", "plan", projectDir})
			})
			if !strings.Contains(output, "recommended steps:") || !strings.Contains(output, "无需升级动作") {
				t.Fatalf("expected no-op upgrade plan output, got:\n%s", output)
			}

			readmePath := filepath.Join(projectDir, "README.md")
			readmeData, err := os.ReadFile(readmePath)
			if err != nil {
				t.Fatalf("read generated README: %v", err)
			}
			if err := os.WriteFile(readmePath, append(readmeData, []byte("\nlocal drift\n")...), 0o644); err != nil {
				t.Fatalf("write local drift README: %v", err)
			}

			output = captureStdout(t, func() error {
				return run([]string{"upgrade", "inspect", projectDir})
			})
			if !strings.Contains(output, "compatibility level: manual_review") || !strings.Contains(output, "local modified files: README.md") {
				t.Fatalf("expected manual_review upgrade inspect output, got:\n%s", output)
			}
		})
	})
}

func TestCLIUpgradeInspectAndPlanRejectMissingMetadata(t *testing.T) {
	t.Setenv("FIBERX_MANIFEST_ROOT", manifestRootForCLI(t))
	workdir := t.TempDir()

	err := run([]string{"upgrade", "inspect", workdir})
	if err == nil || !strings.Contains(err.Error(), ".fiberx/manifest.json") {
		t.Fatalf("expected upgrade inspect missing metadata error, got %v", err)
	}

	err = run([]string{"upgrade", "plan", workdir})
	if err == nil || !strings.Contains(err.Error(), ".fiberx/manifest.json") {
		t.Fatalf("expected upgrade plan missing metadata error, got %v", err)
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

	var buffer bytes.Buffer
	copyDone := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&buffer, reader)
		copyDone <- copyErr
	}()

	runErr := fn()
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	if err := <-copyDone; err != nil {
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

func copyDir(t *testing.T, sourceDir, targetDir string) {
	t.Helper()

	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, 0o644)
	}); err != nil {
		t.Fatalf("copy directory %q to %q: %v", sourceDir, targetDir, err)
	}
}

func withGeneratorIdentityForCLI(t *testing.T, generatorVersion, generatorCommit string, fn func()) {
	t.Helper()

	previousVersion := version.Version
	previousCommit := version.Commit
	version.Version = generatorVersion
	version.Commit = generatorCommit
	defer func() {
		version.Version = previousVersion
		version.Commit = previousCommit
	}()

	fn()
}
