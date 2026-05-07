package build

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GoFurry/fiberx/internal/buildconfig"
	generatorversion "github.com/GoFurry/fiberx/internal/version"
)

type Options struct {
	TargetNames    []string
	PlatformFilter string
	Clean          bool
	DryRun         bool
	Profile        string
	NoHooks        bool
	AutoApprove    bool
}

type Result struct {
	ProjectDir          string
	OutDir              string
	Profile             string
	Version             VersionInfo
	Artifacts           []Artifact
	ChecksumPath        string
	BuildMetadataPath   string
	ReleaseManifestPath string
	DryRun              bool
	HooksSkipped        bool
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
	ChecksumSHA256    string
	SizeBytes         int64
	PreHooks          []string
	PostHooks         []string
	HooksSkipped      bool
	UPXEnabled        bool
	UPXLevel          int
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
	buildMetadataPath := filepath.Join(outDir, "build-metadata.json")
	releaseManifestPath := filepath.Join(outDir, "release-manifest.json")
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

	artifacts, err := executeTasks(projectDir, cfg, versionInfo, tasks, opts)
	if err != nil {
		return Result{}, err
	}

	sortArtifacts(artifacts)

	checksumPath := ""
	if !opts.DryRun {
		enrichedArtifacts, err := enrichArtifacts(artifacts)
		if err != nil {
			return Result{}, err
		}
		artifacts = enrichedArtifacts
	}
	if cfg.Build.Checksum.Enabled {
		checksumPath = filepath.Join(outDir, "SHA256SUMS")
		if !opts.DryRun {
			if err := writeChecksums(outDir, checksumPath, artifacts); err != nil {
				return Result{}, err
			}
		}
	}
	if !opts.DryRun {
		if err := writeBuildMetadata(outDir, buildMetadataPath, releaseManifestPath, cfg, opts, versionInfo, artifacts, checksumPath); err != nil {
			return Result{}, err
		}
		if err := writeReleaseManifest(outDir, releaseManifestPath, cfg, opts, versionInfo, artifacts, checksumPath); err != nil {
			return Result{}, err
		}
	}

	return Result{
		ProjectDir:          projectDir,
		OutDir:              outDir,
		Profile:             opts.Profile,
		Version:             versionInfo,
		Artifacts:           artifacts,
		ChecksumPath:        checksumPath,
		BuildMetadataPath:   buildMetadataPath,
		ReleaseManifestPath: releaseManifestPath,
		DryRun:              opts.DryRun,
		HooksSkipped:        opts.NoHooks && artifactsHaveHooks(artifacts),
	}, nil
}

func executeTasks(projectDir string, cfg buildconfig.File, versionInfo VersionInfo, tasks []buildTask, opts Options) ([]Artifact, error) {
	if len(tasks) == 0 {
		return []Artifact{}, nil
	}

	if !cfg.Build.Parallel {
		artifacts := make([]Artifact, 0, len(tasks))
		for _, task := range tasks {
			artifact, err := executeTask(projectDir, cfg, versionInfo, task, opts)
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
					artifact, err := executeTask(projectDir, cfg, versionInfo, task, opts)
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

func executeTask(projectDir string, cfg buildconfig.File, versionInfo VersionInfo, task buildTask, opts Options) (Artifact, error) {
	artifact := Artifact{
		TargetName:        task.Target.Name,
		Package:           task.Target.Package,
		Platform:          platformLabel(task.Platform),
		OutputPath:        task.OutputPath,
		ArchivePath:       task.ArchivePath,
		DistributablePath: task.DistributablePath,
		PreHooks:          hookNames(task.Target.PreHooks),
		PostHooks:         hookNames(task.Target.PostHooks),
		HooksSkipped:      opts.NoHooks && (len(task.Target.PreHooks) > 0 || len(task.Target.PostHooks) > 0),
		UPXEnabled:        cfg.Build.Compress.UPX.Enabled,
		UPXLevel:          cfg.Build.Compress.UPX.Level,
	}
	if !opts.DryRun {
		if err := os.MkdirAll(filepath.Dir(task.OutputPath), 0o755); err != nil {
			return Artifact{}, fmt.Errorf("create artifact directory for %q: %w", task.OutputPath, err)
		}
		if !opts.NoHooks {
			if err := runHooks(projectDir, cfg.Build.OutDir, task, opts.Profile, "pre", task.Target.PreHooks); err != nil {
				return Artifact{}, err
			}
		}
		if err := runGoBuild(projectDir, cfg, task.Target, task.Platform, task.OutputPath, versionInfo); err != nil {
			return Artifact{}, err
		}
		if cfg.Build.Compress.UPX.Enabled {
			if err := runUPX(task.OutputPath, cfg.Build.Compress.UPX.Level); err != nil {
				return Artifact{}, fmt.Errorf("compress target %q for %s with upx: %w", task.Target.Name, platformLabel(task.Platform), err)
			}
		}
		if !opts.NoHooks {
			if err := runHooks(projectDir, cfg.Build.OutDir, task, opts.Profile, "post", task.Target.PostHooks); err != nil {
				return Artifact{}, err
			}
		}
		if task.ArchivePath != "" {
			if err := writeArchive(projectDir, task, task.Target.Archive.Format); err != nil {
				return Artifact{}, err
			}
		}
	}
	return artifact, nil
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
		relativePath, err := filepath.Rel(outDir, distributablePath)
		if err != nil {
			return fmt.Errorf("relativize checksum path %q: %w", distributablePath, err)
		}
		lines = append(lines, artifact.ChecksumSHA256+"  "+filepath.ToSlash(relativePath))
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

func enrichArtifacts(artifacts []Artifact) ([]Artifact, error) {
	enriched := make([]Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		distributablePath := artifact.DistributablePath
		if distributablePath == "" {
			distributablePath = artifact.OutputPath
		}
		hashValue, err := hashFileSHA256(distributablePath)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(distributablePath)
		if err != nil {
			return nil, fmt.Errorf("inspect distributable %q: %w", distributablePath, err)
		}
		artifact.ChecksumSHA256 = hashValue
		artifact.SizeBytes = info.Size()
		enriched = append(enriched, artifact)
	}
	return enriched, nil
}

type buildMetadataFile struct {
	SchemaVersion   string                  `json:"schema_version"`
	GeneratedAt     string                  `json:"generated_at"`
	Project         buildMetadataProject    `json:"project"`
	Generator       buildMetadataGenerator  `json:"generator"`
	Build           buildMetadataBuild      `json:"build"`
	Version         VersionInfo             `json:"version"`
	Artifacts       []buildMetadataArtifact `json:"artifacts"`
	ChecksumFile    string                  `json:"checksum_file,omitempty"`
	ReleaseManifest string                  `json:"release_manifest"`
}

type buildMetadataProject struct {
	Name   string `json:"name"`
	Module string `json:"module"`
}

type buildMetadataGenerator struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type buildMetadataBuild struct {
	Profile       string `json:"profile,omitempty"`
	Clean         bool   `json:"clean"`
	Parallel      bool   `json:"parallel"`
	VersionSource string `json:"version_source"`
	HooksSkipped  bool   `json:"hooks_skipped"`
	UPXEnabled    bool   `json:"upx_enabled"`
	UPXLevel      int    `json:"upx_level"`
}

type buildMetadataArtifact struct {
	Target            string   `json:"target"`
	Package           string   `json:"package"`
	Platform          string   `json:"platform"`
	BinaryPath        string   `json:"binary_path"`
	ArchivePath       string   `json:"archive_path,omitempty"`
	DistributablePath string   `json:"distributable_path"`
	ChecksumSHA256    string   `json:"checksum_sha256"`
	PreHooks          []string `json:"pre_hooks,omitempty"`
	PostHooks         []string `json:"post_hooks,omitempty"`
	HooksSkipped      bool     `json:"hooks_skipped"`
	UPXEnabled        bool     `json:"upx_enabled"`
	UPXLevel          int      `json:"upx_level,omitempty"`
}

type releaseManifestFile struct {
	SchemaVersion string                    `json:"schema_version"`
	Project       buildMetadataProject      `json:"project"`
	Profile       string                    `json:"profile,omitempty"`
	GeneratedAt   string                    `json:"generated_at"`
	Version       VersionInfo               `json:"version"`
	Artifacts     []releaseManifestArtifact `json:"artifacts"`
	Checksums     releaseManifestChecksums  `json:"checksums"`
}

type releaseManifestArtifact struct {
	Target            string   `json:"target"`
	Platform          string   `json:"platform"`
	DistributablePath string   `json:"distributable_path"`
	ArchiveEnabled    bool     `json:"archive_enabled"`
	ChecksumSHA256    string   `json:"checksum_sha256"`
	SizeBytes         int64    `json:"size_bytes"`
	HooksSkipped      bool     `json:"hooks_skipped"`
	UPXEnabled        bool     `json:"upx_enabled"`
	HooksApplied      []string `json:"hooks_applied,omitempty"`
}

type releaseManifestChecksums struct {
	File      string `json:"file,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
}

func writeBuildMetadata(outDir, metadataPath, releaseManifestPath string, cfg buildconfig.File, opts Options, versionInfo VersionInfo, artifacts []Artifact, checksumPath string) error {
	payload := buildMetadataFile{
		SchemaVersion: "v1",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Project: buildMetadataProject{
			Name:   cfg.Project.Name,
			Module: cfg.Project.Module,
		},
		Generator: buildMetadataGenerator{
			Version: generatorversion.Version,
			Commit:  generatorversion.Commit,
		},
		Build: buildMetadataBuild{
			Profile:       opts.Profile,
			Clean:         opts.Clean || cfg.Build.Clean,
			Parallel:      cfg.Build.Parallel,
			VersionSource: cfg.Build.Version.Source,
			HooksSkipped:  opts.NoHooks && artifactsHaveHooks(artifacts),
			UPXEnabled:    cfg.Build.Compress.UPX.Enabled,
			UPXLevel:      cfg.Build.Compress.UPX.Level,
		},
		Version:         versionInfo,
		Artifacts:       make([]buildMetadataArtifact, 0, len(artifacts)),
		ChecksumFile:    relSlash(outDir, checksumPath),
		ReleaseManifest: relSlash(outDir, releaseManifestPath),
	}
	for _, artifact := range artifacts {
		payload.Artifacts = append(payload.Artifacts, buildMetadataArtifact{
			Target:            artifact.TargetName,
			Package:           artifact.Package,
			Platform:          artifact.Platform,
			BinaryPath:        relSlash(outDir, artifact.OutputPath),
			ArchivePath:       relSlash(outDir, artifact.ArchivePath),
			DistributablePath: relSlash(outDir, distributablePath(artifact)),
			ChecksumSHA256:    artifact.ChecksumSHA256,
			PreHooks:          append([]string(nil), artifact.PreHooks...),
			PostHooks:         append([]string(nil), artifact.PostHooks...),
			HooksSkipped:      artifact.HooksSkipped,
			UPXEnabled:        artifact.UPXEnabled,
			UPXLevel:          artifact.UPXLevel,
		})
	}
	return writeJSONFile(metadataPath, payload)
}

func writeReleaseManifest(outDir, manifestPath string, cfg buildconfig.File, opts Options, versionInfo VersionInfo, artifacts []Artifact, checksumPath string) error {
	payload := releaseManifestFile{
		SchemaVersion: "v1",
		Project: buildMetadataProject{
			Name:   cfg.Project.Name,
			Module: cfg.Project.Module,
		},
		Profile:     opts.Profile,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Version:     versionInfo,
		Artifacts:   make([]releaseManifestArtifact, 0, len(artifacts)),
		Checksums: releaseManifestChecksums{
			File:      relSlash(outDir, checksumPath),
			Algorithm: checksumAlgorithm(cfg),
		},
	}
	for _, artifact := range artifacts {
		payload.Artifacts = append(payload.Artifacts, releaseManifestArtifact{
			Target:            artifact.TargetName,
			Platform:          artifact.Platform,
			DistributablePath: relSlash(outDir, distributablePath(artifact)),
			ArchiveEnabled:    artifact.ArchivePath != "",
			ChecksumSHA256:    artifact.ChecksumSHA256,
			SizeBytes:         artifact.SizeBytes,
			HooksSkipped:      artifact.HooksSkipped,
			UPXEnabled:        artifact.UPXEnabled,
			HooksApplied:      appliedHookNames(artifact),
		})
	}
	return writeJSONFile(manifestPath, payload)
}

func artifactsHaveHooks(artifacts []Artifact) bool {
	for _, artifact := range artifacts {
		if len(artifact.PreHooks) > 0 || len(artifact.PostHooks) > 0 {
			return true
		}
	}
	return false
}

func appliedHookNames(artifact Artifact) []string {
	if artifact.HooksSkipped {
		return nil
	}
	return append(append([]string{}, artifact.PreHooks...), artifact.PostHooks...)
}

func writeJSONFile(path string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", path, err)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %q: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}

func relSlash(baseDir, path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	relative, err := filepath.Rel(baseDir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}

func distributablePath(artifact Artifact) string {
	if artifact.DistributablePath != "" {
		return artifact.DistributablePath
	}
	return artifact.OutputPath
}

func checksumAlgorithm(cfg buildconfig.File) string {
	if !cfg.Build.Checksum.Enabled {
		return ""
	}
	return cfg.Build.Checksum.Algorithm
}

func hookNames(hooks []buildconfig.Hook) []string {
	names := make([]string, 0, len(hooks))
	for _, hook := range hooks {
		name := strings.TrimSpace(hook.Name)
		if name == "" {
			name = strings.Join(hook.Command, " ")
		}
		names = append(names, name)
	}
	return names
}

func runHooks(projectDir, outDir string, task buildTask, profile string, stage string, hooks []buildconfig.Hook) error {
	for index, hook := range hooks {
		hookName := strings.TrimSpace(hook.Name)
		if hookName == "" {
			hookName = stage + "-" + strconv.Itoa(index+1)
		}
		command := exec.Command(hook.Command[0], hook.Command[1:]...)
		command.Dir = projectDir
		if strings.TrimSpace(hook.Dir) != "" {
			command.Dir = filepath.Join(projectDir, filepath.FromSlash(hook.Dir))
		}
		env := append(os.Environ(),
			"FIBERX_BUILD_TARGET="+task.Target.Name,
			"FIBERX_BUILD_GOOS="+task.Platform.GOOS,
			"FIBERX_BUILD_GOARCH="+task.Platform.GOARCH,
			"FIBERX_BUILD_PROFILE="+profile,
			"FIBERX_BUILD_PROJECT_DIR="+projectDir,
			"FIBERX_BUILD_OUT_DIR="+filepath.Join(projectDir, outDir),
			"FIBERX_BUILD_BINARY="+task.OutputPath,
			"FIBERX_BUILD_ARCHIVE="+task.ArchivePath,
			"FIBERX_BUILD_DISTRIBUTABLE="+task.DistributablePath,
		)
		for key, value := range hook.Env {
			env = append(env, key+"="+value)
		}
		command.Env = env
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s hook %q for target %q on %s failed: %s", stage, hookName, task.Target.Name, platformLabel(task.Platform), strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func runUPX(binaryPath string, level int) error {
	upxPath, err := exec.LookPath("upx")
	if err != nil {
		return fmt.Errorf("upx was not found in PATH")
	}
	cmd := exec.Command(upxPath, fmt.Sprintf("-%d", level), binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
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
