package buildconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadParsesMinimalSupportedBuildConfig(t *testing.T) {
	projectDir := t.TempDir()
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  out_dir: dist
  clean: true
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
    ldflags:
      - "-s -w"
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
        - windows/amd64
`)

	cfg, err := Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Project.Name != "demo" || cfg.Project.Module != "github.com/example/demo" {
		t.Fatalf("unexpected project config: %#v", cfg.Project)
	}
	if cfg.Build.Version.Source != "git" {
		t.Fatalf("expected version source git, got %q", cfg.Build.Version.Source)
	}
	if len(cfg.Build.Targets) != 1 || cfg.Build.Targets[0].Package != "." {
		t.Fatalf("unexpected targets: %#v", cfg.Build.Targets)
	}
}

func TestLoadRejectsMissingConfig(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), Filename) {
		t.Fatalf("expected missing config error, got %v", err)
	}
}

func TestLoadRejectsUnsupportedPhase15P0Fields(t *testing.T) {
	projectDir := t.TempDir()
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  parallel: true
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
	if err == nil || !strings.Contains(err.Error(), "build.parallel") {
		t.Fatalf("expected unsupported field error, got %v", err)
	}
}

func TestLoadRejectsInvalidPlatform(t *testing.T) {
	projectDir := t.TempDir()
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

func writeBuildConfig(t *testing.T, projectDir, contents string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(projectDir, Filename), []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("write build config: %v", err)
	}
}
