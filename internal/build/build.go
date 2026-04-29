package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/GoFurry/fiberx/internal/buildconfig"
)

type Options struct {
	TargetNames    []string
	PlatformFilter string
	Clean          bool
}

type Result struct {
	ProjectDir string
	OutDir     string
	Version    VersionInfo
	Artifacts  []Artifact
}

type VersionInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

type Artifact struct {
	TargetName string
	Package    string
	Platform   string
	OutputPath string
}

type Platform struct {
	GOOS   string
	GOARCH string
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

	if opts.Clean || cfg.Build.Clean {
		if err := os.RemoveAll(outDir); err != nil {
			return Result{}, fmt.Errorf("clean output directory %q: %w", outDir, err)
		}
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", outDir, err)
	}

	filterPlatform := strings.TrimSpace(strings.ToLower(opts.PlatformFilter))
	artifacts := make([]Artifact, 0)
	for _, target := range selectedTargets {
		targetPlatforms, err := selectPlatforms(target, filterPlatform)
		if err != nil {
			return Result{}, err
		}

		for _, platform := range targetPlatforms {
			outputPath := outputPathForTarget(outDir, target, platform)
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return Result{}, fmt.Errorf("create artifact directory for %q: %w", outputPath, err)
			}
			if err := validatePackage(projectDir, target.Package); err != nil {
				return Result{}, err
			}
			if err := runGoBuild(projectDir, cfg, target, platform, outputPath, versionInfo); err != nil {
				return Result{}, err
			}
			artifacts = append(artifacts, Artifact{
				TargetName: target.Name,
				Package:    target.Package,
				Platform:   platformLabel(platform),
				OutputPath: outputPath,
			})
		}
	}

	return Result{
		ProjectDir: projectDir,
		OutDir:     outDir,
		Version:    versionInfo,
		Artifacts:  artifacts,
	}, nil
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
