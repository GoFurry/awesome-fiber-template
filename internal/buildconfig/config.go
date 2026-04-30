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
	OutDir   string             `yaml:"out_dir"`
	Clean    bool               `yaml:"clean"`
	Parallel bool               `yaml:"parallel"`
	Version  Version            `yaml:"version"`
	Defaults Defaults           `yaml:"defaults"`
	Targets  []Target           `yaml:"targets"`
	Checksum Checksum           `yaml:"checksum"`
	Compress Compress           `yaml:"compress"`
	Profiles map[string]Profile `yaml:"profiles"`
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
	PreHooks  []Hook   `yaml:"pre_hooks"`
	PostHooks []Hook   `yaml:"post_hooks"`
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

type Compress struct {
	UPX UPX `yaml:"upx"`
}

type UPX struct {
	Enabled bool `yaml:"enabled"`
	Level   int  `yaml:"level"`
}

type Hook struct {
	Name    string            `yaml:"name"`
	Command []string          `yaml:"command"`
	Dir     string            `yaml:"dir"`
	Env     map[string]string `yaml:"env"`
}

type Profile struct {
	OutDir   string          `yaml:"out_dir"`
	Clean    *bool           `yaml:"clean"`
	Parallel *bool           `yaml:"parallel"`
	Defaults ProfileDefaults `yaml:"defaults"`
	Targets  []ProfileTarget `yaml:"targets"`
	Checksum ProfileChecksum `yaml:"checksum"`
}

type ProfileDefaults struct {
	CGO      *bool    `yaml:"cgo"`
	TrimPath *bool    `yaml:"trimpath"`
	Ldflags  []string `yaml:"ldflags"`
}

type ProfileChecksum struct {
	Enabled   *bool  `yaml:"enabled"`
	Algorithm string `yaml:"algorithm"`
}

type ProfileTarget struct {
	Name      string         `yaml:"name"`
	Output    string         `yaml:"output"`
	Platforms []string       `yaml:"platforms"`
	Archive   ProfileArchive `yaml:"archive"`
}

type ProfileArchive struct {
	Enabled *bool    `yaml:"enabled"`
	Format  string   `yaml:"format"`
	Files   []string `yaml:"files"`
}

func Load(projectDir string) (File, error) {
	return LoadWithProfile(projectDir, "")
}

func LoadWithProfile(projectDir, profileName string) (File, error) {
	profileName = strings.TrimSpace(profileName)
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
	if err := validateConfig(projectDir, cfg, "Phase 15 P3-M1"); err != nil {
		return File{}, err
	}
	if err := validateProfiles(projectDir, cfg); err != nil {
		return File{}, err
	}
	if strings.TrimSpace(profileName) == "" {
		return cfg, nil
	}

	applied, err := applyProfile(cfg, profileName)
	if err != nil {
		return File{}, err
	}
	if err := validateConfig(projectDir, applied, "Phase 15 P3-M1"); err != nil {
		return File{}, err
	}
	return applied, nil
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
		normalizeHooks(f.Build.Targets[index].PreHooks)
		normalizeHooks(f.Build.Targets[index].PostHooks)
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
	if len(f.Build.Profiles) > 0 {
		normalizedProfiles := make(map[string]Profile, len(f.Build.Profiles))
		for name, profile := range f.Build.Profiles {
			profile.OutDir = strings.TrimSpace(profile.OutDir)
			profile.Checksum.Algorithm = strings.TrimSpace(strings.ToLower(profile.Checksum.Algorithm))
			if profile.Defaults.Ldflags != nil {
				profile.Defaults.Ldflags = trimStrings(profile.Defaults.Ldflags)
			}
			for index := range profile.Targets {
				profile.Targets[index].Name = strings.TrimSpace(profile.Targets[index].Name)
				profile.Targets[index].Output = strings.TrimSpace(profile.Targets[index].Output)
				profile.Targets[index].Archive.Format = strings.TrimSpace(strings.ToLower(profile.Targets[index].Archive.Format))
				if profile.Targets[index].Platforms != nil {
					profile.Targets[index].Platforms = trimLowerStrings(profile.Targets[index].Platforms)
				}
				if profile.Targets[index].Archive.Files != nil {
					profile.Targets[index].Archive.Files = trimStrings(profile.Targets[index].Archive.Files)
				}
			}
			normalizedProfiles[strings.TrimSpace(name)] = profile
		}
		f.Build.Profiles = normalizedProfiles
	}
	if f.Build.Compress.UPX.Level == 0 {
		f.Build.Compress.UPX.Level = 5
	}
}

func validateUnsupported(root yaml.Node) error {
	unsupportedPaths := []string{
		"build.pre_hooks",
		"build.post_hooks",
	}
	for _, path := range unsupportedPaths {
		if hasPath(root, path) {
			return fmt.Errorf("build config field %q is not supported in Phase 15 P3-M1", path)
		}
	}
	if err := validateUnsupportedProfiles(root); err != nil {
		return err
	}
	if err := validateUnsupportedCompress(root); err != nil {
		return err
	}
	return nil
}

func validateConfig(projectDir string, cfg File, phaseLabel string) error {
	if strings.TrimSpace(cfg.Project.Name) == "" {
		return fmt.Errorf("build config field %q is required", "project.name")
	}
	if strings.TrimSpace(cfg.Project.Module) == "" {
		return fmt.Errorf("build config field %q is required", "project.module")
	}
	if cfg.Build.Version.Source != "git" {
		return fmt.Errorf("build version source %q is not supported in %s", cfg.Build.Version.Source, phaseLabel)
	}
	if strings.TrimSpace(cfg.Build.Version.Package) == "" {
		return fmt.Errorf("build config field %q is required", "build.version.package")
	}
	if cfg.Build.Checksum.Algorithm != "sha256" {
		return fmt.Errorf("build checksum algorithm %q is not supported in %s", cfg.Build.Checksum.Algorithm, phaseLabel)
	}
	if cfg.Build.Compress.UPX.Level < 1 || cfg.Build.Compress.UPX.Level > 9 {
		return fmt.Errorf("build upx level %d is not supported in %s", cfg.Build.Compress.UPX.Level, phaseLabel)
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
			return fmt.Errorf("build target %q archive format %q is not supported in %s", target.Name, target.Archive.Format, phaseLabel)
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
		for _, hookSet := range [][]Hook{target.PreHooks, target.PostHooks} {
			for _, hook := range hookSet {
				if len(hook.Command) == 0 {
					return fmt.Errorf("build target %q hook command must contain at least one element", target.Name)
				}
				for _, part := range hook.Command {
					if strings.TrimSpace(part) == "" {
						return fmt.Errorf("build target %q hook command elements must not be empty", target.Name)
					}
				}
				if err := validateHookDir(projectDir, hook.Dir, target.Name); err != nil {
					return err
				}
				for key := range hook.Env {
					if strings.TrimSpace(key) == "" {
						return fmt.Errorf("build target %q hook env keys must not be empty", target.Name)
					}
				}
			}
		}
	}

	return nil
}

func validateProfiles(projectDir string, cfg File) error {
	for profileName, profile := range cfg.Build.Profiles {
		if profileName == "" {
			return fmt.Errorf("build profile names must not be empty")
		}
		if profile.Checksum.Algorithm != "" && profile.Checksum.Algorithm != "sha256" {
			return fmt.Errorf("build profile %q checksum algorithm %q is not supported in Phase 15 P3-M1", profileName, profile.Checksum.Algorithm)
		}
		baseTargets := map[string]Target{}
		for _, target := range cfg.Build.Targets {
			baseTargets[target.Name] = target
		}
		seenTargets := map[string]bool{}
		for _, patch := range profile.Targets {
			if patch.Name == "" {
				return fmt.Errorf("build profile %q target field %q is required", profileName, "name")
			}
			if seenTargets[patch.Name] {
				return fmt.Errorf("build profile %q target %q is duplicated", profileName, patch.Name)
			}
			seenTargets[patch.Name] = true
			_, ok := baseTargets[patch.Name]
			if !ok {
				return fmt.Errorf("build profile %q target patch %q does not match any base target", profileName, patch.Name)
			}
			if patch.Archive.Format != "" && patch.Archive.Format != "auto" && patch.Archive.Format != "zip" && patch.Archive.Format != "tar.gz" {
				return fmt.Errorf("build profile %q target %q archive format %q is not supported in Phase 15 P3-M1", profileName, patch.Name, patch.Archive.Format)
			}
			for _, platform := range patch.Platforms {
				if _, _, ok := parsePlatform(platform); !ok {
					return fmt.Errorf("build profile %q target %q platform %q must use goos/goarch format", profileName, patch.Name, platform)
				}
			}
			for _, archiveFile := range patch.Archive.Files {
				if archiveFile == "" {
					return fmt.Errorf("build profile %q target %q archive file path must not be empty", profileName, patch.Name)
				}
			}
		}
		applied, err := applyProfile(cfg, profileName)
		if err != nil {
			return err
		}
		if err := validateConfig(projectDir, applied, "Phase 15 P3-M1"); err != nil {
			return err
		}
	}
	return nil
}

func applyProfile(cfg File, profileName string) (File, error) {
	profile, ok := cfg.Build.Profiles[profileName]
	if !ok {
		return File{}, fmt.Errorf("build profile %q was not found in %s", profileName, Filename)
	}

	applied := cfg
	if profile.OutDir != "" {
		applied.Build.OutDir = profile.OutDir
	}
	if profile.Clean != nil {
		applied.Build.Clean = *profile.Clean
	}
	if profile.Parallel != nil {
		applied.Build.Parallel = *profile.Parallel
	}
	if profile.Defaults.CGO != nil {
		applied.Build.Defaults.CGO = *profile.Defaults.CGO
	}
	if profile.Defaults.TrimPath != nil {
		applied.Build.Defaults.TrimPath = *profile.Defaults.TrimPath
	}
	if profile.Defaults.Ldflags != nil {
		applied.Build.Defaults.Ldflags = append([]string(nil), profile.Defaults.Ldflags...)
	}
	if profile.Checksum.Enabled != nil {
		applied.Build.Checksum.Enabled = *profile.Checksum.Enabled
	}
	if profile.Checksum.Algorithm != "" {
		applied.Build.Checksum.Algorithm = profile.Checksum.Algorithm
	}

	indexByName := make(map[string]int, len(applied.Build.Targets))
	for index, target := range applied.Build.Targets {
		indexByName[target.Name] = index
	}
	for _, patch := range profile.Targets {
		index, ok := indexByName[patch.Name]
		if !ok {
			return File{}, fmt.Errorf("build profile %q target patch %q does not match any base target", profileName, patch.Name)
		}
		target := applied.Build.Targets[index]
		if patch.Output != "" {
			target.Output = patch.Output
		}
		if patch.Platforms != nil {
			target.Platforms = append([]string(nil), patch.Platforms...)
		}
		if patch.Archive.Enabled != nil {
			target.Archive.Enabled = *patch.Archive.Enabled
		}
		if patch.Archive.Format != "" {
			target.Archive.Format = patch.Archive.Format
		}
		if patch.Archive.Files != nil {
			target.Archive.Files = append([]string(nil), patch.Archive.Files...)
		}
		applied.Build.Targets[index] = target
	}

	applied.Build.Profiles = cfg.Build.Profiles
	return applied, nil
}

func validateUnsupportedProfiles(root yaml.Node) error {
	buildNode, ok := mappingValue(firstContent(&root), "build")
	if !ok {
		return nil
	}
	profilesNode, ok := mappingValue(buildNode, "profiles")
	if !ok || profilesNode.Kind != yaml.MappingNode {
		return nil
	}

	for index := 0; index+1 < len(profilesNode.Content); index += 2 {
		profileName := profilesNode.Content[index].Value
		profileNode := profilesNode.Content[index+1]
		for path, label := range map[string]string{
			"project":              "build.profiles.<name>.project",
			"version":              "build.profiles.<name>.version",
			"compress":             "build.profiles.<name>.compress",
			"pre_hooks":            "build.profiles.<name>.pre_hooks",
			"post_hooks":           "build.profiles.<name>.post_hooks",
			"targets[].package":    "build.profiles.<name>.targets[].package",
			"targets[].pre_hooks":  "build.profiles.<name>.targets[].pre_hooks",
			"targets[].post_hooks": "build.profiles.<name>.targets[].post_hooks",
		} {
			if hasPathNode(profileNode, strings.Split(path, ".")) {
				return fmt.Errorf("build config field %q is not supported in Phase 15 P3-M1", strings.ReplaceAll(label, "<name>", profileName))
			}
		}
	}
	return nil
}

func validateUnsupportedCompress(root yaml.Node) error {
	buildNode, ok := mappingValue(firstContent(&root), "build")
	if !ok {
		return nil
	}
	compressNode, ok := mappingValue(buildNode, "compress")
	if !ok {
		return nil
	}
	if compressNode.Kind != yaml.MappingNode {
		return fmt.Errorf("build config field %q is not supported in Phase 15 P3-M2", "build.compress")
	}
	for index := 0; index+1 < len(compressNode.Content); index += 2 {
		key := compressNode.Content[index].Value
		if key != "upx" {
			return fmt.Errorf("build config field %q is not supported in Phase 15 P3-M2", "build.compress."+key)
		}
		upxNode := compressNode.Content[index+1]
		if upxNode.Kind != yaml.MappingNode {
			return fmt.Errorf("build config field %q is not supported in Phase 15 P3-M2", "build.compress.upx")
		}
		for childIndex := 0; childIndex+1 < len(upxNode.Content); childIndex += 2 {
			childKey := upxNode.Content[childIndex].Value
			if childKey != "enabled" && childKey != "level" {
				return fmt.Errorf("build config field %q is not supported in Phase 15 P3-M2", "build.compress.upx."+childKey)
			}
		}
	}
	return nil
}

func firstContent(node *yaml.Node) *yaml.Node {
	if node == nil || len(node.Content) == 0 {
		return nil
	}
	return node.Content[0]
}

func trimStrings(values []string) []string {
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		trimmed = append(trimmed, strings.TrimSpace(value))
	}
	return trimmed
}

func trimLowerStrings(values []string) []string {
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		trimmed = append(trimmed, strings.TrimSpace(strings.ToLower(value)))
	}
	return trimmed
}

func normalizeHooks(hooks []Hook) {
	for index := range hooks {
		hooks[index].Name = strings.TrimSpace(hooks[index].Name)
		hooks[index].Dir = strings.TrimSpace(hooks[index].Dir)
		for commandIndex := range hooks[index].Command {
			hooks[index].Command[commandIndex] = strings.TrimSpace(hooks[index].Command[commandIndex])
		}
		if hooks[index].Env == nil {
			continue
		}
		normalizedEnv := make(map[string]string, len(hooks[index].Env))
		for key, value := range hooks[index].Env {
			normalizedEnv[strings.TrimSpace(key)] = value
		}
		hooks[index].Env = normalizedEnv
	}
}

func validateHookDir(projectDir, dir, targetName string) error {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	if filepath.IsAbs(dir) {
		return fmt.Errorf("build target %q hook dir %q must be relative to the project root", targetName, dir)
	}
	resolvedPath := filepath.Clean(filepath.Join(projectDir, filepath.FromSlash(dir)))
	projectRoot := filepath.Clean(projectDir)
	relToProject, err := filepath.Rel(projectRoot, resolvedPath)
	if err != nil || strings.HasPrefix(relToProject, "..") {
		return fmt.Errorf("build target %q hook dir %q must stay within the project root", targetName, dir)
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("build target %q hook dir %q does not exist", targetName, dir)
		}
		return fmt.Errorf("inspect build target %q hook dir %q: %w", targetName, dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("build target %q hook dir %q must be a directory", targetName, dir)
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
