package buildconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadParsesPhase15P2BuildConfig(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  out_dir: dist
  clean: true
  parallel: true
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
    ldflags:
      - "-s -w"
  checksum:
    enabled: true
    algorithm: sha256
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
        - windows/amd64
      archive:
        enabled: true
        format: auto
        files:
          - README.md
          - config
`)

	cfg, err := Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Project.Name != "demo" || cfg.Project.Module != "github.com/example/demo" {
		t.Fatalf("unexpected project config: %#v", cfg.Project)
	}
	if !cfg.Build.Parallel {
		t.Fatalf("expected parallel=true, got %#v", cfg.Build)
	}
	if !cfg.Build.Checksum.Enabled || cfg.Build.Checksum.Algorithm != "sha256" {
		t.Fatalf("unexpected checksum config: %#v", cfg.Build.Checksum)
	}
	if len(cfg.Build.Targets) != 1 || cfg.Build.Targets[0].Package != "." {
		t.Fatalf("unexpected targets: %#v", cfg.Build.Targets)
	}
	if !cfg.Build.Targets[0].Archive.Enabled || cfg.Build.Targets[0].Archive.Format != "auto" {
		t.Fatalf("unexpected archive config: %#v", cfg.Build.Targets[0].Archive)
	}
}

func TestLoadRejectsMissingConfig(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), Filename) {
		t.Fatalf("expected missing config error, got %v", err)
	}
}

func TestLoadRejectsStillUnsupportedPhase15Fields(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  compress:
    upx:
      enabled: true
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "build.compress") {
		t.Fatalf("expected unsupported field error, got %v", err)
	}
}

func TestLoadRejectsInvalidPlatform(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "goos/goarch") {
		t.Fatalf("expected invalid platform error, got %v", err)
	}
}

func TestLoadRejectsInvalidChecksumAlgorithm(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
  checksum:
    enabled: true
    algorithm: md5
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "checksum algorithm") {
		t.Fatalf("expected checksum algorithm error, got %v", err)
	}
}

func TestLoadRejectsInvalidArchiveFormat(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
      archive:
        enabled: true
        format: rar
        files:
          - README.md
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "archive format") {
		t.Fatalf("expected archive format error, got %v", err)
	}
}

func TestLoadRejectsMissingArchiveFile(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
      archive:
        enabled: true
        files:
          - missing.txt
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing archive file error, got %v", err)
	}
}

func mustWriteProjectSupportFiles(t *testing.T, projectDir string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(projectDir, "README.md"), []byte("demo\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}
	configDir := filepath.Join(projectDir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte("mode: debug\n"), 0o644); err != nil {
		t.Fatalf("write config/app.yaml: %v", err)
	}
}

func writeBuildConfig(t *testing.T, projectDir, contents string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(projectDir, Filename), []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write build config: %v", err)
	}
}
