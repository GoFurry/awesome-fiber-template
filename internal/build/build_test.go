package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/GoFurry/fiberx/internal/buildconfig"
)

func TestExecuteBuildsSelectedTargetsAndPlatforms(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	result, err := Execute(projectDir, cfg, Options{
		TargetNames:    []string{"server"},
		PlatformFilter: runtime.GOOS + "/" + runtime.GOARCH,
	})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if len(result.Artifacts) != 1 {
		t.Fatalf("expected one artifact, got %#v", result.Artifacts)
	}
	artifact := result.Artifacts[0]
	if artifact.TargetName != "server" {
		t.Fatalf("expected target server, got %#v", artifact)
	}
	if _, err := os.Stat(artifact.OutputPath); err != nil {
		t.Fatalf("expected built artifact at %q: %v", artifact.OutputPath, err)
	}
}

func TestExecuteCleanRemovesPreviousOutputs(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	staleDir := filepath.Join(projectDir, "dist", "stale")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatalf("create stale dir: %v", err)
	}
	staleFile := filepath.Join(staleDir, "old.txt")
	if err := os.WriteFile(staleFile, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	if _, err := Execute(projectDir, cfg, Options{
		TargetNames:    []string{"server"},
		PlatformFilter: runtime.GOOS + "/" + runtime.GOARCH,
		Clean:          true,
	}); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, got %v", err)
	}
}

func TestExecuteRejectsMissingPackage(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	cfg.Build.Targets[0].Package = "./cmd/missing"

	_, err = Execute(projectDir, cfg, Options{
		TargetNames:    []string{"server"},
		PlatformFilter: runtime.GOOS + "/" + runtime.GOARCH,
	})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing package error, got %v", err)
	}
}

func TestExecuteFailsWithoutGitMetadata(t *testing.T) {
	projectDir := buildProjectFixtureWithoutGit(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	_, err = Execute(projectDir, cfg, Options{
		TargetNames:    []string{"server"},
		PlatformFilter: runtime.GOOS + "/" + runtime.GOARCH,
	})
	if err == nil || !strings.Contains(err.Error(), "resolve build version from git") {
		t.Fatalf("expected git metadata error, got %v", err)
	}
}

func TestRenderLdflags(t *testing.T) {
	ldflags := renderLdflags([]string{
		"-s -w",
		"-X {{.VersionPackage}}.Version={{.Version}}",
		"-X {{.VersionPackage}}.Commit={{.Commit}}",
		"-X {{.VersionPackage}}.BuildTime={{.BuildTime}}",
	}, "github.com/example/demo/internal/version", VersionInfo{
		Version:   "v1.2.3",
		Commit:    "abc123",
		BuildTime: "2026-04-29T00:00:00Z",
	})

	if !strings.Contains(ldflags, "github.com/example/demo/internal/version.Version=v1.2.3") {
		t.Fatalf("expected rendered version ldflag, got %q", ldflags)
	}
	if !strings.Contains(ldflags, "Commit=abc123") || !strings.Contains(ldflags, "BuildTime=2026-04-29T00:00:00Z") {
		t.Fatalf("expected rendered commit/buildtime ldflags, got %q", ldflags)
	}
}

func buildProjectFixture(t *testing.T) string {
	t.Helper()

	projectDir := buildProjectFixtureWithoutGit(t)
	runCommand(t, projectDir, "git", "init")
	runCommand(t, projectDir, "git", "config", "user.name", "fiberx-test")
	runCommand(t, projectDir, "git", "config", "user.email", "fiberx@example.com")
	runCommand(t, projectDir, "git", "add", ".")
	runCommand(t, projectDir, "git", "commit", "-m", "init")
	runCommand(t, projectDir, "git", "tag", "v0.1.0")
	return projectDir
}

func buildProjectFixtureWithoutGit(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), `module github.com/example/demo

go 1.26.0
`)
	writeFile(t, filepath.Join(projectDir, "main.go"), `package main

import projectversion "github.com/example/demo/internal/version"

func main() {
	_, _, _ = projectversion.Version, projectversion.Commit, projectversion.BuildTime
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "version", "version.go"), `package version

var (
	Version = "dev"
	Commit = "unknown"
	BuildTime = ""
)
`)
	buildConfig := `project:
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
      - "-X {{.VersionPackage}}.Version={{.Version}}"
      - "-X {{.VersionPackage}}.Commit={{.Commit}}"
      - "-X {{.VersionPackage}}.BuildTime={{.BuildTime}}"
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - ` + runtime.GOOS + `/` + runtime.GOARCH + `
`
	writeFile(t, filepath.Join(projectDir, buildconfig.Filename), buildConfig)
	return projectDir
}

func runCommand(t *testing.T, dir string, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, string(output))
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
