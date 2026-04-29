package buildconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const Filename = "fiberx.yaml"

type File struct {
	Project Project `yaml:"project"`
	Build   Build   `yaml:"build"`
}

type Project struct {
	Name   string `yaml:"name"`
	Module string `yaml:"module"`
}

type Build struct {
	OutDir   string   `yaml:"out_dir"`
	Clean    bool     `yaml:"clean"`
	Parallel bool     `yaml:"parallel"`
	Version  Version  `yaml:"version"`
	Defaults Defaults `yaml:"defaults"`
	Targets  []Target `yaml:"targets"`
	Checksum Checksum `yaml:"checksum"`
}

type Version struct {
	Source  string `yaml:"source"`
	Package string `yaml:"package"`
}

type Defaults struct {
	CGO      bool     `yaml:"cgo"`
	TrimPath bool     `yaml:"trimpath"`
	Ldflags  []string `yaml:"ldflags"`
}

type Target struct {
	Name      string   `yaml:"name"`
	Package   string   `yaml:"package"`
	Output    string   `yaml:"output"`
	Platforms []string `yaml:"platforms"`
	Archive   Archive  `yaml:"archive"`
}

type Archive struct {
	Enabled bool     `yaml:"enabled"`
	Format  string   `yaml:"format"`
	Files   []string `yaml:"files"`
}

type Checksum struct {
	Enabled   bool   `yaml:"enabled"`
	Algorithm string `yaml:"algorithm"`
}

func Load(projectDir string) (File, error) {
	configPath := filepath.Join(projectDir, Filename)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, fmt.Errorf("build config %q was not found", Filename)
		}
		return File{}, fmt.Errorf("read build config %q: %w", configPath, err)
	}

	var raw yaml.Node
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return File{}, fmt.Errorf("decode build config %q: %w", configPath, err)
	}

	var cfg File
	if err := raw.Decode(&cfg); err != nil {
		return File{}, fmt.Errorf("decode build config %q: %w", configPath, err)
	}

	cfg.normalize()
	if err := validateUnsupported(raw); err != nil {
		return File{}, err
	}
	if err := validateConfig(projectDir, cfg); err != nil {
		return File{}, err
	}

	return cfg, nil
}

func (f *File) normalize() {
	if strings.TrimSpace(f.Build.OutDir) == "" {
		f.Build.OutDir = "dist"
	}
	f.Build.Version.Source = strings.TrimSpace(strings.ToLower(f.Build.Version.Source))
	f.Build.Checksum.Algorithm = strings.TrimSpace(strings.ToLower(f.Build.Checksum.Algorithm))
	if f.Build.Checksum.Algorithm == "" {
		f.Build.Checksum.Algorithm = "sha256"
	}
	for index := range f.Build.Targets {
		f.Build.Targets[index].Name = strings.TrimSpace(f.Build.Targets[index].Name)
		f.Build.Targets[index].Package = strings.TrimSpace(f.Build.Targets[index].Package)
		f.Build.Targets[index].Output = strings.TrimSpace(f.Build.Targets[index].Output)
		f.Build.Targets[index].Archive.Format = strings.TrimSpace(strings.ToLower(f.Build.Targets[index].Archive.Format))
		if f.Build.Targets[index].Archive.Format == "" {
			f.Build.Targets[index].Archive.Format = "auto"
		}
		for platformIndex := range f.Build.Targets[index].Platforms {
			f.Build.Targets[index].Platforms[platformIndex] = strings.TrimSpace(strings.ToLower(f.Build.Targets[index].Platforms[platformIndex]))
		}
		for fileIndex := range f.Build.Targets[index].Archive.Files {
			f.Build.Targets[index].Archive.Files[fileIndex] = strings.TrimSpace(f.Build.Targets[index].Archive.Files[fileIndex])
		}
	}
}

func validateUnsupported(root yaml.Node) error {
	unsupportedPaths := []string{
		"build.compress",
		"build.profiles",
		"build.targets[].pre_hooks",
		"build.targets[].post_hooks",
	}
	for _, path := range unsupportedPaths {
		if hasPath(root, path) {
			return fmt.Errorf("build config field %q is not supported in Phase 15 P0", path)
		}
	}
	return nil
}

func validateConfig(projectDir string, cfg File) error {
	if strings.TrimSpace(cfg.Project.Name) == "" {
		return fmt.Errorf("build config field %q is required", "project.name")
	}
	if strings.TrimSpace(cfg.Project.Module) == "" {
		return fmt.Errorf("build config field %q is required", "project.module")
	}
	if cfg.Build.Version.Source != "git" {
		return fmt.Errorf("build version source %q is not supported in Phase 15 P0", cfg.Build.Version.Source)
	}
	if strings.TrimSpace(cfg.Build.Version.Package) == "" {
		return fmt.Errorf("build config field %q is required", "build.version.package")
	}
	if cfg.Build.Checksum.Algorithm != "sha256" {
		return fmt.Errorf("build checksum algorithm %q is not supported in Phase 15 P2", cfg.Build.Checksum.Algorithm)
	}
	if len(cfg.Build.Targets) == 0 {
		return fmt.Errorf("build config field %q must contain at least one target", "build.targets")
	}

	seen := map[string]bool{}
	for _, target := range cfg.Build.Targets {
		if target.Name == "" {
			return fmt.Errorf("build target field %q is required", "name")
		}
		if seen[target.Name] {
			return fmt.Errorf("build target %q is duplicated", target.Name)
		}
		seen[target.Name] = true
		if target.Package == "" {
			return fmt.Errorf("build target %q field %q is required", target.Name, "package")
		}
		if target.Output == "" {
			return fmt.Errorf("build target %q field %q is required", target.Name, "output")
		}
		if len(target.Platforms) == 0 {
			return fmt.Errorf("build target %q field %q must contain at least one platform", target.Name, "platforms")
		}
		if target.Archive.Format != "auto" && target.Archive.Format != "zip" && target.Archive.Format != "tar.gz" {
			return fmt.Errorf("build target %q archive format %q is not supported in Phase 15 P2", target.Name, target.Archive.Format)
		}
		for _, platform := range target.Platforms {
			if _, _, ok := parsePlatform(platform); !ok {
				return fmt.Errorf("build target %q platform %q must use goos/goarch format", target.Name, platform)
			}
		}
		for _, archiveFile := range target.Archive.Files {
			if archiveFile == "" {
				return fmt.Errorf("build target %q archive file path must not be empty", target.Name)
			}
			if filepath.IsAbs(archiveFile) {
				return fmt.Errorf("build target %q archive file %q must be relative to the project root", target.Name, archiveFile)
			}
			resolvedPath := filepath.Clean(filepath.Join(projectDir, filepath.FromSlash(archiveFile)))
			projectRoot := filepath.Clean(projectDir)
			relToProject, err := filepath.Rel(projectRoot, resolvedPath)
			if err != nil || strings.HasPrefix(relToProject, "..") {
				return fmt.Errorf("build target %q archive file %q must stay within the project root", target.Name, archiveFile)
			}
			info, err := os.Stat(resolvedPath)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("build target %q archive file %q does not exist", target.Name, archiveFile)
				}
				return fmt.Errorf("inspect build target %q archive file %q: %w", target.Name, archiveFile, err)
			}
			if !(info.Mode().IsRegular() || info.IsDir()) {
				return fmt.Errorf("build target %q archive file %q must be a regular file or directory", target.Name, archiveFile)
			}
			outDirPath := filepath.Clean(filepath.Join(projectDir, cfg.Build.OutDir))
			relToOutDir, err := filepath.Rel(outDirPath, resolvedPath)
			if err == nil && relToOutDir != "." && !strings.HasPrefix(relToOutDir, "..") {
				return fmt.Errorf("build target %q archive file %q must not point inside out_dir %q", target.Name, archiveFile, cfg.Build.OutDir)
			}
			if resolvedPath == outDirPath {
				return fmt.Errorf("build target %q archive file %q must not point to out_dir %q", target.Name, archiveFile, cfg.Build.OutDir)
			}
		}
	}

	return nil
}

func parsePlatform(raw string) (string, string, bool) {
	parts := strings.Split(strings.TrimSpace(raw), "/")
	if len(parts) != 2 {
		return "", "", false
	}
	goos := strings.TrimSpace(parts[0])
	goarch := strings.TrimSpace(parts[1])
	if goos == "" || goarch == "" {
		return "", "", false
	}
	return goos, goarch, true
}

func hasPath(root yaml.Node, path string) bool {
	if len(root.Content) == 0 {
		return false
	}
	return hasPathNode(root.Content[0], strings.Split(path, "."))
}

func hasPathNode(node *yaml.Node, parts []string) bool {
	if len(parts) == 0 {
		return true
	}

	part := parts[0]
	if strings.HasSuffix(part, "[]") {
		key := strings.TrimSuffix(part, "[]")
		child, ok := mappingValue(node, key)
		if !ok || child.Kind != yaml.SequenceNode {
			return false
		}
		for _, item := range child.Content {
			if hasPathNode(item, parts[1:]) {
				return true
			}
		}
		return false
	}

	child, ok := mappingValue(node, part)
	if !ok {
		return false
	}
	return hasPathNode(child, parts[1:])
}

func mappingValue(node *yaml.Node, key string) (*yaml.Node, bool) {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil, false
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1], true
		}
	}
	return nil, false
}
