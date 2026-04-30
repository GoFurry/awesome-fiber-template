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

func TestLoadWithProfileAppliesOverlay(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  out_dir: dist
  clean: true
  parallel: false
  version:
    source: git
    package: github.com/example/demo/internal/version
  defaults:
    cgo: false
    trimpath: true
    ldflags:
      - "-s -w"
  checksum:
    enabled: false
    algorithm: sha256
  profiles:
    prod:
      out_dir: dist/prod
      parallel: true
      checksum:
        enabled: true
        algorithm: sha256
      targets:
        - name: server
          output: demo-prod
          platforms:
            - linux/amd64
          archive:
            enabled: true
            format: zip
            files:
              - README.md
              - config
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
        - windows/amd64
      archive:
        enabled: false
        format: auto
        files:
          - README.md
          - config
`)

	cfg, err := LoadWithProfile(projectDir, "prod")
	if err != nil {
		t.Fatalf("LoadWithProfile() returned error: %v", err)
	}
	if cfg.Build.OutDir != "dist/prod" || !cfg.Build.Parallel || !cfg.Build.Checksum.Enabled {
		t.Fatalf("expected profile overlay on build config, got %#v", cfg.Build)
	}
	target := cfg.Build.Targets[0]
	if target.Output != "demo-prod" || len(target.Platforms) != 1 || target.Platforms[0] != "linux/amd64" {
		t.Fatalf("expected target overlay to apply, got %#v", target)
	}
	if !target.Archive.Enabled || target.Archive.Format != "zip" {
		t.Fatalf("expected archive overlay to apply, got %#v", target.Archive)
	}
}

func TestLoadParsesTargetHooksAndUPX(t *testing.T) {
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
  compress:
    upx:
      enabled: true
      level: 7
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
      pre_hooks:
        - name: generate
          command: ["go", "generate", "./..."]
          dir: "."
          env:
            MODE: test
      post_hooks:
        - command: ["go", "version"]
`)

	cfg, err := Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if !cfg.Build.Compress.UPX.Enabled || cfg.Build.Compress.UPX.Level != 7 {
		t.Fatalf("unexpected upx config: %#v", cfg.Build.Compress.UPX)
	}
	if len(cfg.Build.Targets[0].PreHooks) != 1 || len(cfg.Build.Targets[0].PostHooks) != 1 {
		t.Fatalf("unexpected hook config: %#v", cfg.Build.Targets[0])
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
  post_hooks:
    - command: ["go", "version"]
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
	if err == nil || !strings.Contains(err.Error(), "build.post_hooks") {
		t.Fatalf("expected unsupported field error, got %v", err)
	}
}

func TestLoadRejectsUnsupportedGlobalHooks(t *testing.T) {
	projectDir := t.TempDir()
	mustWriteProjectSupportFiles(t, projectDir)
	writeBuildConfig(t, projectDir, `
project:
  name: demo
  module: github.com/example/demo
build:
  pre_hooks:
    - command: ["go", "version"]
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
	if err == nil || !strings.Contains(err.Error(), "build.pre_hooks") {
		t.Fatalf("expected unsupported global hooks error, got %v", err)
	}
}

func TestLoadRejectsMissingProfile(t *testing.T) {
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
`)

	_, err := LoadWithProfile(projectDir, "prod")
	if err == nil || !strings.Contains(err.Error(), `build profile "prod" was not found`) {
		t.Fatalf("expected missing profile error, got %v", err)
	}
}

func TestLoadRejectsUnsupportedProfileFields(t *testing.T) {
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
  profiles:
    prod:
      compress:
        upx:
          enabled: true
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "build.profiles.prod.compress") {
		t.Fatalf("expected unsupported profile field error, got %v", err)
	}
}

func TestLoadRejectsProfileUnknownTargetPatch(t *testing.T) {
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
  profiles:
    prod:
      targets:
        - name: worker
          output: worker
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), `target patch "worker" does not match any base target`) {
		t.Fatalf("expected unknown target patch error, got %v", err)
	}
}

func TestLoadRejectsProfileTargetHooks(t *testing.T) {
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
  profiles:
    prod:
      targets:
        - name: server
          pre_hooks:
            - command: ["go", "version"]
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "build.profiles.prod.targets[].pre_hooks") {
		t.Fatalf("expected unsupported profile target hooks error, got %v", err)
	}
}

func TestLoadRejectsEmptyHookCommand(t *testing.T) {
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
      pre_hooks:
        - name: empty
          command: []
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "hook command must contain at least one element") {
		t.Fatalf("expected empty hook command error, got %v", err)
	}
}

func TestLoadRejectsHookDirOutsideProject(t *testing.T) {
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
      pre_hooks:
        - command: ["go", "version"]
          dir: "../outside"
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "must stay within the project root") {
		t.Fatalf("expected hook dir boundary error, got %v", err)
	}
}

func TestLoadRejectsInvalidUPXLevel(t *testing.T) {
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
  compress:
    upx:
      enabled: true
      level: 10
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - linux/amd64
`)

	_, err := Load(projectDir)
	if err == nil || !strings.Contains(err.Error(), "upx level 10") {
		t.Fatalf("expected invalid upx level error, got %v", err)
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
