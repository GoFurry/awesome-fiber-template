package build

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
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

func TestExecuteDryRunPlansArtifactsWithoutWritingOutputs(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	cfg.Build.Targets[0].Archive.Enabled = true
	cfg.Build.Targets[0].Archive.Format = "auto"
	cfg.Build.Checksum.Enabled = true

	result, err := Execute(projectDir, cfg, Options{
		TargetNames:    []string{"server"},
		PlatformFilter: runtime.GOOS + "/" + runtime.GOARCH,
		DryRun:         true,
	})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if !result.DryRun {
		t.Fatalf("expected dry-run result, got %#v", result)
	}
	if len(result.Artifacts) != 1 {
		t.Fatalf("expected one planned artifact, got %#v", result.Artifacts)
	}
	artifact := result.Artifacts[0]
	if artifact.ArchivePath == "" || artifact.DistributablePath == "" {
		t.Fatalf("expected planned archive/distributable paths, got %#v", artifact)
	}
	if _, err := os.Stat(result.OutDir); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to create out dir, got %v", err)
	}
	if _, err := os.Stat(result.ChecksumPath); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to create checksum file, got %v", err)
	}
	if _, err := os.Stat(result.BuildMetadataPath); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to create build metadata file, got %v", err)
	}
	if _, err := os.Stat(result.ReleaseManifestPath); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to create release manifest file, got %v", err)
	}
}

func TestExecuteWritesArchiveAndChecksums(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	cfg.Build.Targets[0].Archive.Enabled = true
	cfg.Build.Targets[0].Archive.Format = "auto"
	cfg.Build.Checksum.Enabled = true

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
	if _, err := os.Stat(artifact.ArchivePath); err != nil {
		t.Fatalf("expected archive at %q: %v", artifact.ArchivePath, err)
	}
	if _, err := os.Stat(result.ChecksumPath); err != nil {
		t.Fatalf("expected checksum file at %q: %v", result.ChecksumPath, err)
	}
	if _, err := os.Stat(result.BuildMetadataPath); err != nil {
		t.Fatalf("expected build metadata file at %q: %v", result.BuildMetadataPath, err)
	}
	if _, err := os.Stat(result.ReleaseManifestPath); err != nil {
		t.Fatalf("expected release manifest file at %q: %v", result.ReleaseManifestPath, err)
	}

	entries := listArchiveEntries(t, artifact.ArchivePath)
	assertContainsAll(t, entries, []string{
		filepath.Base(artifact.OutputPath),
		"README.md",
		filepath.ToSlash(filepath.Join("config", "app.yaml")),
	})

	checksumData, err := os.ReadFile(result.ChecksumPath)
	if err != nil {
		t.Fatalf("read checksum file: %v", err)
	}
	if !strings.Contains(string(checksumData), filepath.ToSlash(filepath.Join("server", runtime.GOOS+"_"+runtime.GOARCH, filepath.Base(artifact.ArchivePath)))) {
		t.Fatalf("expected checksum file to reference archive, got:\n%s", string(checksumData))
	}

	metadataPayload := struct {
		Profile         string `json:"profile"`
		ReleaseManifest string `json:"release_manifest"`
		Artifacts       []struct {
			Target         string `json:"target"`
			ChecksumSHA256 string `json:"checksum_sha256"`
		} `json:"artifacts"`
	}{}
	metadataData, err := os.ReadFile(result.BuildMetadataPath)
	if err != nil {
		t.Fatalf("read build metadata: %v", err)
	}
	if err := json.Unmarshal(metadataData, &metadataPayload); err != nil {
		t.Fatalf("unmarshal build metadata: %v", err)
	}
	if metadataPayload.ReleaseManifest != "release-manifest.json" || len(metadataPayload.Artifacts) != 1 || metadataPayload.Artifacts[0].ChecksumSHA256 == "" {
		t.Fatalf("unexpected build metadata payload: %#v", metadataPayload)
	}

	releasePayload := struct {
		Checksums struct {
			File      string `json:"file"`
			Algorithm string `json:"algorithm"`
		} `json:"checksums"`
		Artifacts []struct {
			Target            string `json:"target"`
			DistributablePath string `json:"distributable_path"`
			ChecksumSHA256    string `json:"checksum_sha256"`
			SizeBytes         int64  `json:"size_bytes"`
		} `json:"artifacts"`
	}{}
	releaseData, err := os.ReadFile(result.ReleaseManifestPath)
	if err != nil {
		t.Fatalf("read release manifest: %v", err)
	}
	if err := json.Unmarshal(releaseData, &releasePayload); err != nil {
		t.Fatalf("unmarshal release manifest: %v", err)
	}
	if releasePayload.Checksums.File != "SHA256SUMS" || releasePayload.Checksums.Algorithm != "sha256" {
		t.Fatalf("unexpected release checksum payload: %#v", releasePayload.Checksums)
	}
	if len(releasePayload.Artifacts) != 1 || releasePayload.Artifacts[0].ChecksumSHA256 == "" || releasePayload.Artifacts[0].SizeBytes <= 0 {
		t.Fatalf("unexpected release manifest artifacts: %#v", releasePayload.Artifacts)
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

func TestArchiveFormatForTarget(t *testing.T) {
	zipFormat, err := archiveFormatForTarget(Platform{GOOS: "windows", GOARCH: "amd64"}, "auto")
	if err != nil {
		t.Fatalf("archiveFormatForTarget windows auto returned error: %v", err)
	}
	if zipFormat != "zip" {
		t.Fatalf("expected windows auto format to be zip, got %q", zipFormat)
	}

	tarFormat, err := archiveFormatForTarget(Platform{GOOS: "linux", GOARCH: "amd64"}, "auto")
	if err != nil {
		t.Fatalf("archiveFormatForTarget linux auto returned error: %v", err)
	}
	if tarFormat != "tar.gz" {
		t.Fatalf("expected linux auto format to be tar.gz, got %q", tarFormat)
	}

	overrideFormat, err := archiveFormatForTarget(Platform{GOOS: "linux", GOARCH: "amd64"}, "zip")
	if err != nil {
		t.Fatalf("archiveFormatForTarget zip override returned error: %v", err)
	}
	if overrideFormat != "zip" {
		t.Fatalf("expected explicit zip override, got %q", overrideFormat)
	}
}

func TestExecuteParallelMatchesSerialArtifacts(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	cfg.Build.Targets[0].Platforms = []string{runtime.GOOS + "/" + runtime.GOARCH, alternatePlatformForBuildTest()}

	serialResult, err := Execute(projectDir, cfg, Options{TargetNames: []string{"server"}})
	if err != nil {
		t.Fatalf("serial Execute() returned error: %v", err)
	}

	cfg.Build.Parallel = true
	parallelResult, err := Execute(projectDir, cfg, Options{
		TargetNames: []string{"server"},
		Clean:       true,
	})
	if err != nil {
		t.Fatalf("parallel Execute() returned error: %v", err)
	}

	if strings.Join(ArtifactPaths(serialResult), "\n") != strings.Join(ArtifactPaths(parallelResult), "\n") {
		t.Fatalf("expected serial and parallel artifact sets to match\nserial=%v\nparallel=%v", ArtifactPaths(serialResult), ArtifactPaths(parallelResult))
	}
}

func TestExecuteAppliesProfileOverlay(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.LoadWithProfile(projectDir, "prod")
	if err != nil {
		t.Fatalf("LoadWithProfile() returned error: %v", err)
	}

	result, err := Execute(projectDir, cfg, Options{
		TargetNames: []string{"server"},
		Profile:     "prod",
	})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if result.Profile != "prod" {
		t.Fatalf("expected profile prod, got %#v", result)
	}
	if !strings.Contains(filepath.ToSlash(result.OutDir), "dist/prod") {
		t.Fatalf("expected profile out_dir override, got %q", result.OutDir)
	}
	if len(result.Artifacts) != 1 {
		t.Fatalf("expected one artifact from prod profile, got %#v", result.Artifacts)
	}
	artifact := result.Artifacts[0]
	if artifact.TargetName != "server" || artifact.ArchivePath == "" {
		t.Fatalf("expected archive-enabled prod artifact, got %#v", artifact)
	}
}

func TestExecuteRunsHooksAndWritesHookMetadata(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	logPath := filepath.Join(projectDir, "hook.log")
	cfg.Build.Targets[0].PreHooks = []buildconfig.Hook{{
		Name:    "pre-log",
		Command: []string{"go", "run", "./cmd/hookprobe", "pre"},
		Env: map[string]string{
			"HOOK_LOG": logPath,
		},
	}}
	cfg.Build.Targets[0].PostHooks = []buildconfig.Hook{{
		Name:    "post-log",
		Command: []string{"go", "run", "./cmd/hookprobe", "post"},
		Env: map[string]string{
			"HOOK_LOG": logPath,
		},
	}}

	result, err := Execute(projectDir, cfg, Options{
		TargetNames: []string{"server"},
	})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read hook log: %v", err)
	}
	if strings.TrimSpace(string(logData)) != "pre\npost" && strings.TrimSpace(string(logData)) != "pre\r\npost" {
		t.Fatalf("expected hook log order pre/post, got %q", string(logData))
	}

	metadataData, err := os.ReadFile(result.BuildMetadataPath)
	if err != nil {
		t.Fatalf("read build metadata: %v", err)
	}
	var metadataPayload struct {
		Artifacts []struct {
			PreHooks  []string `json:"pre_hooks"`
			PostHooks []string `json:"post_hooks"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(metadataData, &metadataPayload); err != nil {
		t.Fatalf("unmarshal build metadata: %v", err)
	}
	if len(metadataPayload.Artifacts) != 1 || len(metadataPayload.Artifacts[0].PreHooks) != 1 || len(metadataPayload.Artifacts[0].PostHooks) != 1 {
		t.Fatalf("expected hook metadata to be recorded, got %#v", metadataPayload)
	}
}

func TestExecuteFailsWhenHookFails(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	cfg.Build.Targets[0].PreHooks = []buildconfig.Hook{{
		Name:    "fail",
		Command: []string{"go", "run", "./cmd/hookprobe", "fail"},
	}}

	_, err = Execute(projectDir, cfg, Options{
		TargetNames: []string{"server"},
	})
	if err == nil || !strings.Contains(err.Error(), `pre hook "fail"`) {
		t.Fatalf("expected failing pre hook error, got %v", err)
	}
}

func TestRunUPXMissing(t *testing.T) {
	originalPath := os.Getenv("PATH")
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("LookPath(go): %v", err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("LookPath(git): %v", err)
	}
	pathEntries := []string{filepath.Dir(goPath)}
	if filepath.Dir(gitPath) != filepath.Dir(goPath) {
		pathEntries = append(pathEntries, filepath.Dir(gitPath))
	}
	t.Setenv("PATH", strings.Join(pathEntries, string(os.PathListSeparator)))
	defer os.Setenv("PATH", originalPath)

	err = runUPX(filepath.Join(t.TempDir(), "binary"), 5)
	if err == nil || !strings.Contains(err.Error(), "upx was not found in PATH") {
		t.Fatalf("expected missing upx error, got %v", err)
	}
}

func TestExecuteRunsUPXWhenEnabled(t *testing.T) {
	projectDir := buildProjectFixture(t)
	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	cfg.Build.Compress.UPX.Enabled = true
	cfg.Build.Compress.UPX.Level = 7
	upxLogPath := filepath.Join(projectDir, "upx.log")
	upxDir := createFakeUPX(t, upxLogPath)
	t.Setenv("PATH", upxDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	result, err := Execute(projectDir, cfg, Options{
		TargetNames: []string{"server"},
	})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	upxLogData, err := os.ReadFile(upxLogPath)
	if err != nil {
		t.Fatalf("read upx log: %v", err)
	}
	if !strings.Contains(string(upxLogData), "-7") {
		t.Fatalf("expected fake upx log to contain level flag, got %q", string(upxLogData))
	}

	releaseData, err := os.ReadFile(result.ReleaseManifestPath)
	if err != nil {
		t.Fatalf("read release manifest: %v", err)
	}
	var releasePayload struct {
		Artifacts []struct {
			UPXEnabled bool `json:"upx_enabled"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(releaseData, &releasePayload); err != nil {
		t.Fatalf("unmarshal release manifest: %v", err)
	}
	if len(releasePayload.Artifacts) != 1 || !releasePayload.Artifacts[0].UPXEnabled {
		t.Fatalf("expected upx_enabled in release manifest, got %#v", releasePayload)
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
	writeFile(t, filepath.Join(projectDir, "README.md"), "demo\n")
	writeFile(t, filepath.Join(projectDir, "config", "app.yaml"), "mode: debug\n")
	writeFile(t, filepath.Join(projectDir, "internal", "version", "version.go"), `package version

var (
	Version = "dev"
	Commit = "unknown"
	BuildTime = ""
)
`)
	writeFile(t, filepath.Join(projectDir, "cmd", "hookprobe", "main.go"), `package main

import (
	"fmt"
	"os"
)

func main() {
	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if mode == "fail" {
		fmt.Fprintln(os.Stderr, "hook failure")
		os.Exit(1)
	}
	logPath := os.Getenv("HOOK_LOG")
	if logPath != "" {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		fmt.Fprintln(file, mode)
	}
}
`)
	buildConfig := `project:
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
      - "-X {{.VersionPackage}}.Version={{.Version}}"
      - "-X {{.VersionPackage}}.Commit={{.Commit}}"
      - "-X {{.VersionPackage}}.BuildTime={{.BuildTime}}"
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
          platforms:
            - ` + runtime.GOOS + `/` + runtime.GOARCH + `
          archive:
            enabled: true
            format: auto
            files:
              - README.md
              - config
  targets:
    - name: server
      package: .
      output: demo
      platforms:
        - ` + runtime.GOOS + `/` + runtime.GOARCH + `
      archive:
        enabled: false
        format: auto
        files:
          - README.md
          - config
`
	writeFile(t, filepath.Join(projectDir, buildconfig.Filename), buildConfig)
	return projectDir
}

func alternatePlatformForBuildTest() string {
	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		return "linux/amd64"
	}
	return "windows/amd64"
}

func listArchiveEntries(t *testing.T, archivePath string) []string {
	t.Helper()

	if strings.HasSuffix(archivePath, ".zip") {
		reader, err := zip.OpenReader(archivePath)
		if err != nil {
			t.Fatalf("open zip archive: %v", err)
		}
		defer reader.Close()
		entries := make([]string, 0, len(reader.File))
		for _, file := range reader.File {
			entries = append(entries, file.Name)
		}
		return entries
	}

	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("open tar.gz archive: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	entries := []string{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read tar.gz archive: %v", err)
		}
		entries = append(entries, header.Name)
	}
	return entries
}

func assertContainsAll(t *testing.T, items []string, want []string) {
	t.Helper()

	for _, expected := range want {
		found := false
		for _, item := range items {
			if item == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected %q in %#v", expected, items)
		}
	}
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

func createFakeUPX(t *testing.T, logPath string) string {
	t.Helper()

	dir := t.TempDir()
	var scriptPath string
	var content string
	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(dir, "upx.cmd")
		content = "@echo off\r\n" +
			fmt.Sprintf("echo %%*>>\"%s\"\r\n", strings.ReplaceAll(logPath, "/", "\\")) +
			"exit /b 0\r\n"
	} else {
		scriptPath = filepath.Join(dir, "upx")
		content = "#!/bin/sh\n" +
			fmt.Sprintf("echo \"$@\" >> \"%s\"\n", logPath) +
			"exit 0\n"
	}
	if err := os.WriteFile(scriptPath, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake upx: %v", err)
	}
	return dir
}
