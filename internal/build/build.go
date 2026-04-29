package build

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GoFurry/fiberx/internal/buildconfig"
)

type Options struct {
	TargetNames    []string
	PlatformFilter string
	Clean          bool
	DryRun         bool
}

type Result struct {
	ProjectDir   string
	OutDir       string
	Version      VersionInfo
	Artifacts    []Artifact
	ChecksumPath string
	DryRun       bool
}

type VersionInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

type Artifact struct {
	TargetName        string
	Package           string
	Platform          string
	OutputPath        string
	ArchivePath       string
	DistributablePath string
}

type Platform struct {
	GOOS   string
	GOARCH string
}

type buildTask struct {
	Target            buildconfig.Target
	Platform          Platform
	OutputPath        string
	ArchivePath       string
	DistributablePath string
}

type archiveEntry struct {
	SourcePath  string
	ArchivePath string
	Info        os.FileInfo
}

func Execute(projectDir string, cfg buildconfig.File, opts Options) (Result, error) {
	outDir := filepath.Join(projectDir, cfg.Build.OutDir)
	selectedTargets, err := selectTargets(cfg, opts.TargetNames)
	if err != nil {
		return Result{}, err
	}

	versionInfo, err := resolveVersionInfo(projectDir, cfg)
	if err != nil {
		return Result{}, err
	}

	for _, target := range selectedTargets {
		if err := validatePackage(projectDir, target.Package); err != nil {
			return Result{}, err
		}
	}

	tasks, err := planTasks(projectDir, outDir, selectedTargets, strings.TrimSpace(strings.ToLower(opts.PlatformFilter)))
	if err != nil {
		return Result{}, err
	}

	if !opts.DryRun {
		if opts.Clean || cfg.Build.Clean {
			if err := os.RemoveAll(outDir); err != nil {
				return Result{}, fmt.Errorf("clean output directory %q: %w", outDir, err)
			}
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return Result{}, fmt.Errorf("create output directory %q: %w", outDir, err)
		}
	}

	artifacts, err := executeTasks(projectDir, cfg, versionInfo, tasks, opts.DryRun)
	if err != nil {
		return Result{}, err
	}

	sortArtifacts(artifacts)

	checksumPath := ""
	if cfg.Build.Checksum.Enabled {
		checksumPath = filepath.Join(outDir, "SHA256SUMS")
		if !opts.DryRun {
			if err := writeChecksums(outDir, checksumPath, artifacts); err != nil {
				return Result{}, err
			}
		}
	}

	return Result{
		ProjectDir:   projectDir,
		OutDir:       outDir,
		Version:      versionInfo,
		Artifacts:    artifacts,
		ChecksumPath: checksumPath,
		DryRun:       opts.DryRun,
	}, nil
}

func executeTasks(projectDir string, cfg buildconfig.File, versionInfo VersionInfo, tasks []buildTask, dryRun bool) ([]Artifact, error) {
	if len(tasks) == 0 {
		return []Artifact{}, nil
	}

	if !cfg.Build.Parallel {
		artifacts := make([]Artifact, 0, len(tasks))
		for _, task := range tasks {
			artifact, err := executeTask(projectDir, cfg, versionInfo, task, dryRun)
			if err != nil {
				return nil, err
			}
			artifacts = append(artifacts, artifact)
		}
		return artifacts, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskCh := make(chan buildTask)
	resultCh := make(chan Artifact, len(tasks))
	errCh := make(chan error, 1)

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}

	var wg sync.WaitGroup
	for index := 0; index < workerCount; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-taskCh:
					if !ok {
						return
					}
					artifact, err := executeTask(projectDir, cfg, versionInfo, task, dryRun)
					if err != nil {
						select {
						case errCh <- err:
						default:
						}
						cancel()
						return
					}
					resultCh <- artifact
				}
			}
		}()
	}

	go func() {
		defer close(taskCh)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case taskCh <- task:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	artifacts := make([]Artifact, 0, len(tasks))
	for artifact := range resultCh {
		artifacts = append(artifacts, artifact)
	}

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return artifacts, nil
}

func executeTask(projectDir string, cfg buildconfig.File, versionInfo VersionInfo, task buildTask, dryRun bool) (Artifact, error) {
	if !dryRun {
		if err := os.MkdirAll(filepath.Dir(task.OutputPath), 0o755); err != nil {
			return Artifact{}, fmt.Errorf("create artifact directory for %q: %w", task.OutputPath, err)
		}
		if err := runGoBuild(projectDir, cfg, task.Target, task.Platform, task.OutputPath, versionInfo); err != nil {
			return Artifact{}, err
		}
		if task.ArchivePath != "" {
			if err := writeArchive(projectDir, task, task.Target.Archive.Format); err != nil {
				return Artifact{}, err
			}
		}
	}

	return Artifact{
		TargetName:        task.Target.Name,
		Package:           task.Target.Package,
		Platform:          platformLabel(task.Platform),
		OutputPath:        task.OutputPath,
		ArchivePath:       task.ArchivePath,
		DistributablePath: task.DistributablePath,
	}, nil
}

func planTasks(projectDir, outDir string, targets []buildconfig.Target, filter string) ([]buildTask, error) {
	tasks := make([]buildTask, 0)
	for _, target := range targets {
		targetPlatforms, err := selectPlatforms(target, filter)
		if err != nil {
			return nil, err
		}
		for _, platform := range targetPlatforms {
			outputPath := outputPathForTarget(outDir, target, platform)
			archivePath, distributablePath, err := distributablePathsForTarget(projectDir, outDir, target, platform, outputPath)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, buildTask{
				Target:            target,
				Platform:          platform,
				OutputPath:        outputPath,
				ArchivePath:       archivePath,
				DistributablePath: distributablePath,
			})
		}
	}
	return tasks, nil
}

func selectTargets(cfg buildconfig.File, names []string) ([]buildconfig.Target, error) {
	if len(names) == 0 {
		return append([]buildconfig.Target(nil), cfg.Build.Targets...), nil
	}

	byName := make(map[string]buildconfig.Target, len(cfg.Build.Targets))
	for _, target := range cfg.Build.Targets {
		byName[target.Name] = target
	}

	selected := make([]buildconfig.Target, 0, len(names))
	for _, name := range names {
		target, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("build target %q was not found in %s", name, buildconfig.Filename)
		}
		selected = append(selected, target)
	}
	return selected, nil
}

func selectPlatforms(target buildconfig.Target, filter string) ([]Platform, error) {
	platforms := make([]Platform, 0, len(target.Platforms))
	for _, raw := range target.Platforms {
		platform, err := parsePlatform(raw)
		if err != nil {
			return nil, fmt.Errorf("build target %q: %w", target.Name, err)
		}
		if filter != "" && platformLabel(platform) != filter {
			continue
		}
		platforms = append(platforms, platform)
	}
	if filter != "" && len(platforms) == 0 {
		return nil, fmt.Errorf("build target %q does not support platform %q", target.Name, filter)
	}
	return platforms, nil
}

func parsePlatform(raw string) (Platform, error) {
	parts := strings.Split(strings.TrimSpace(raw), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return Platform{}, fmt.Errorf("platform %q must use goos/goarch format", raw)
	}
	return Platform{
		GOOS:   strings.TrimSpace(parts[0]),
		GOARCH: strings.TrimSpace(parts[1]),
	}, nil
}

func resolveVersionInfo(projectDir string, cfg buildconfig.File) (VersionInfo, error) {
	if cfg.Build.Version.Source != "git" {
		return VersionInfo{}, fmt.Errorf("build version source %q is not supported in Phase 15 P0", cfg.Build.Version.Source)
	}

	version, err := runGit(projectDir, "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return VersionInfo{}, fmt.Errorf("resolve build version from git: %w", err)
	}
	commit, err := runGit(projectDir, "rev-parse", "--short", "HEAD")
	if err != nil {
		return VersionInfo{}, fmt.Errorf("resolve build commit from git: %w", err)
	}

	return VersionInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func runGit(projectDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", strings.Join(append([]string{"git"}, args...), " "), strings.TrimSpace(string(output)))
	}
	value := strings.TrimSpace(string(output))
	if value == "" {
		return "", fmt.Errorf("%s returned an empty value", strings.Join(append([]string{"git"}, args...), " "))
	}
	return value, nil
}

func validatePackage(projectDir, pkg string) error {
	if pkg != "." && !strings.HasPrefix(pkg, "./cmd/") {
		return fmt.Errorf("build target package %q is not supported in Phase 15 P0", pkg)
	}
	if pkg != "." {
		packageDir := filepath.Join(projectDir, filepath.FromSlash(strings.TrimPrefix(pkg, "./")))
		info, err := os.Stat(packageDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("build target package %q does not exist", pkg)
			}
			return fmt.Errorf("inspect build target package %q: %w", pkg, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("build target package %q is not a directory", pkg)
		}
	}

	cmd := exec.Command("go", "list", pkg)
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("validate build target package %q: %s", pkg, strings.TrimSpace(string(output)))
	}
	return nil
}

func runGoBuild(projectDir string, cfg buildconfig.File, target buildconfig.Target, platform Platform, outputPath string, versionInfo VersionInfo) error {
	args := []string{"build"}
	if cfg.Build.Defaults.TrimPath {
		args = append(args, "-trimpath")
	}
	ldflags := renderLdflags(cfg.Build.Defaults.Ldflags, cfg.Build.Version.Package, versionInfo)
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	args = append(args, "-o", outputPath, target.Package)

	cmd := exec.Command("go", args...)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"GOOS="+platform.GOOS,
		"GOARCH="+platform.GOARCH,
		"CGO_ENABLED="+boolToEnabled(cfg.Build.Defaults.CGO),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build target %q for %s: %s", target.Name, platformLabel(platform), strings.TrimSpace(string(output)))
	}
	return nil
}

func renderLdflags(flags []string, versionPackage string, versionInfo VersionInfo) string {
	if len(flags) == 0 {
		return ""
	}

	rendered := make([]string, 0, len(flags))
	for _, flag := range flags {
		value := flag
		replacements := map[string]string{
			"{{.VersionPackage}}": versionPackage,
			"{{.Version}}":        versionInfo.Version,
			"{{.Commit}}":         versionInfo.Commit,
			"{{.BuildTime}}":      versionInfo.BuildTime,
		}
		for placeholder, replacement := range replacements {
			value = strings.ReplaceAll(value, placeholder, replacement)
		}
		rendered = append(rendered, value)
	}
	return strings.Join(rendered, " ")
}

func outputPathForTarget(outDir string, target buildconfig.Target, platform Platform) string {
	binary := target.Output
	if platform.GOOS == "windows" {
		binary += ".exe"
	}
	return filepath.Join(outDir, target.Name, platform.GOOS+"_"+platform.GOARCH, binary)
}

func distributablePathsForTarget(projectDir, outDir string, target buildconfig.Target, platform Platform, outputPath string) (string, string, error) {
	if !target.Archive.Enabled {
		return "", outputPath, nil
	}

	format, err := archiveFormatForTarget(platform, target.Archive.Format)
	if err != nil {
		return "", "", fmt.Errorf("build target %q: %w", target.Name, err)
	}
	archiveName := target.Output + "_" + platform.GOOS + "_" + platform.GOARCH
	switch format {
	case "zip":
		archiveName += ".zip"
	case "tar.gz":
		archiveName += ".tar.gz"
	default:
		return "", "", fmt.Errorf("build target %q: unsupported archive format %q", target.Name, format)
	}

	archivePath := filepath.Join(outDir, target.Name, platform.GOOS+"_"+platform.GOARCH, archiveName)
	for _, archiveFile := range target.Archive.Files {
		resolvedPath := filepath.Clean(filepath.Join(projectDir, filepath.FromSlash(archiveFile)))
		if resolvedPath == filepath.Clean(outDir) {
			return "", "", fmt.Errorf("build target %q archive file %q must not point to out_dir %q", target.Name, archiveFile, outDir)
		}
	}
	return archivePath, archivePath, nil
}

func archiveFormatForTarget(platform Platform, format string) (string, error) {
	switch format {
	case "", "auto":
		if platform.GOOS == "windows" {
			return "zip", nil
		}
		return "tar.gz", nil
	case "zip", "tar.gz":
		return format, nil
	default:
		return "", fmt.Errorf("archive format %q is not supported in Phase 15 P2", format)
	}
}

func writeArchive(projectDir string, task buildTask, requestedFormat string) error {
	format, err := archiveFormatForTarget(task.Platform, requestedFormat)
	if err != nil {
		return fmt.Errorf("build target %q: %w", task.Target.Name, err)
	}

	entries, err := collectArchiveEntries(projectDir, task)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(task.ArchivePath), 0o755); err != nil {
		return fmt.Errorf("create archive directory for %q: %w", task.ArchivePath, err)
	}

	switch format {
	case "zip":
		return writeZipArchive(task.Target.Name, task.Platform, task.ArchivePath, entries)
	case "tar.gz":
		return writeTarGzArchive(task.Target.Name, task.Platform, task.ArchivePath, entries)
	default:
		return fmt.Errorf("build target %q: unsupported archive format %q", task.Target.Name, format)
	}
}

func collectArchiveEntries(projectDir string, task buildTask) ([]archiveEntry, error) {
	entries := []archiveEntry{}
	seen := map[string]bool{}

	addEntry := func(sourcePath, archivePath string, info os.FileInfo) error {
		normalizedArchivePath := filepath.ToSlash(strings.TrimPrefix(filepath.Clean(archivePath), string(filepath.Separator)))
		if normalizedArchivePath == "." || normalizedArchivePath == "" {
			return fmt.Errorf("build target %q archive path %q is invalid", task.Target.Name, archivePath)
		}
		if seen[normalizedArchivePath] {
			return fmt.Errorf("build target %q archive entry %q is duplicated", task.Target.Name, normalizedArchivePath)
		}
		seen[normalizedArchivePath] = true
		entries = append(entries, archiveEntry{
			SourcePath:  sourcePath,
			ArchivePath: normalizedArchivePath,
			Info:        info,
		})
		return nil
	}

	binaryInfo, err := os.Stat(task.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("inspect built artifact %q: %w", task.OutputPath, err)
	}
	if err := addEntry(task.OutputPath, filepath.Base(task.OutputPath), binaryInfo); err != nil {
		return nil, err
	}

	for _, archiveFile := range task.Target.Archive.Files {
		resolvedPath := filepath.Clean(filepath.Join(projectDir, filepath.FromSlash(archiveFile)))
		info, err := os.Stat(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("inspect build target %q archive file %q: %w", task.Target.Name, archiveFile, err)
		}
		if info.IsDir() {
			err = filepath.Walk(resolvedPath, func(path string, walkInfo os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if walkInfo.IsDir() {
					return nil
				}
				relativeInsideDir, err := filepath.Rel(resolvedPath, path)
				if err != nil {
					return err
				}
				return addEntry(path, filepath.ToSlash(filepath.Join(archiveFile, relativeInsideDir)), walkInfo)
			})
			if err != nil {
				return nil, fmt.Errorf("walk build target %q archive directory %q: %w", task.Target.Name, archiveFile, err)
			}
			continue
		}
		if err := addEntry(resolvedPath, filepath.ToSlash(archiveFile), info); err != nil {
			return nil, err
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ArchivePath < entries[j].ArchivePath
	})
	return entries, nil
}

func writeZipArchive(targetName string, platform Platform, archivePath string, entries []archiveEntry) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive for target %q on %s: %w", targetName, platformLabel(platform), err)
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)

	for _, entry := range entries {
		header, err := zip.FileInfoHeader(entry.Info)
		if err != nil {
			return fmt.Errorf("create zip header for %q: %w", entry.ArchivePath, err)
		}
		header.Name = entry.ArchivePath
		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create zip entry %q: %w", entry.ArchivePath, err)
		}
		if err := copyFile(writer, entry.SourcePath); err != nil {
			return err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("finalize zip archive %q: %w", archivePath, err)
	}
	return nil
}

func writeTarGzArchive(targetName string, platform Platform, archivePath string, entries []archiveEntry) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive for target %q on %s: %w", targetName, platformLabel(platform), err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, entry := range entries {
		header, err := tar.FileInfoHeader(entry.Info, "")
		if err != nil {
			_ = tarWriter.Close()
			_ = gzipWriter.Close()
			return fmt.Errorf("create tar header for %q: %w", entry.ArchivePath, err)
		}
		header.Name = entry.ArchivePath
		if err := tarWriter.WriteHeader(header); err != nil {
			_ = tarWriter.Close()
			_ = gzipWriter.Close()
			return fmt.Errorf("write tar header for %q: %w", entry.ArchivePath, err)
		}
		if err := copyFile(tarWriter, entry.SourcePath); err != nil {
			_ = tarWriter.Close()
			_ = gzipWriter.Close()
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		_ = gzipWriter.Close()
		return fmt.Errorf("finalize tar archive %q: %w", archivePath, err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("finalize gzip archive %q: %w", archivePath, err)
	}
	return nil
}

func copyFile(writer io.Writer, sourcePath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open archive source %q: %w", sourcePath, err)
	}
	defer file.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("copy archive source %q: %w", sourcePath, err)
	}
	return nil
}

func writeChecksums(outDir, checksumPath string, artifacts []Artifact) error {
	lines := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		distributablePath := artifact.DistributablePath
		if distributablePath == "" {
			distributablePath = artifact.OutputPath
		}
		hashValue, err := hashFileSHA256(distributablePath)
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(outDir, distributablePath)
		if err != nil {
			return fmt.Errorf("relativize checksum path %q: %w", distributablePath, err)
		}
		lines = append(lines, hashValue+"  "+filepath.ToSlash(relativePath))
	}
	slices.Sort(lines)
	if err := os.MkdirAll(filepath.Dir(checksumPath), 0o755); err != nil {
		return fmt.Errorf("create checksum directory for %q: %w", checksumPath, err)
	}
	return os.WriteFile(checksumPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func hashFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open distributable %q for checksum: %w", path, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("hash distributable %q: %w", path, err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func sortArtifacts(artifacts []Artifact) {
	sort.SliceStable(artifacts, func(i, j int) bool {
		if artifacts[i].TargetName != artifacts[j].TargetName {
			return artifacts[i].TargetName < artifacts[j].TargetName
		}
		if artifacts[i].Platform != artifacts[j].Platform {
			return artifacts[i].Platform < artifacts[j].Platform
		}
		return artifacts[i].OutputPath < artifacts[j].OutputPath
	})
}

func platformLabel(platform Platform) string {
	return platform.GOOS + "/" + platform.GOARCH
}

func boolToEnabled(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func ArtifactPaths(result Result) []string {
	paths := make([]string, 0, len(result.Artifacts))
	for _, artifact := range result.Artifacts {
		paths = append(paths, artifact.OutputPath)
	}
	slices.Sort(paths)
	return paths
}

func HostPlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}
