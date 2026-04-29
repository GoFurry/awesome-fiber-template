package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/GoFurry/fiberx/internal/build"
	"github.com/GoFurry/fiberx/internal/buildconfig"
	"github.com/GoFurry/fiberx/internal/core"
	"github.com/GoFurry/fiberx/internal/manifest"
	"github.com/GoFurry/fiberx/internal/metadata"
	"github.com/GoFurry/fiberx/internal/report"
	"github.com/GoFurry/fiberx/internal/stack"
	"github.com/GoFurry/fiberx/internal/upgrade"
	"github.com/GoFurry/fiberx/internal/validator"
	"github.com/GoFurry/fiberx/internal/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return nil
	}

	switch args[0] {
	case "new":
		return runNew(args[1:])
	case "init":
		return runInit(args[1:])
	case "list":
		return runList(args[1:])
	case "explain":
		return runExplain(args[1:])
	case "inspect":
		return runInspect(args[1:])
	case "diff":
		return runDiff(args[1:])
	case "upgrade":
		return runUpgrade(args[1:])
	case "build":
		return runBuild(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runNew(args []string) error {
	fs := newFlagSet("new")
	modulePath := fs.String("module", "", "go module path")
	preset := fs.String("preset", "light", "preset name")
	with := fs.String("with", "", "comma-separated capability names")
	fiberVersion := fs.String("fiber-version", stack.DefaultFiberVersion(), "fiber version: v3 or v2")
	cliStyle := fs.String("cli-style", stack.DefaultCLIStyle(), "cli style: cobra or native")
	loggerBackend := fs.String("logger", stack.DefaultLogger(), "logger backend: zap or slog")
	dbKind := fs.String("db", stack.DefaultDB(), "database kind: sqlite, pgsql, or mysql")
	dataAccess := fs.String("data-access", stack.DefaultDataAccess(), "data access stack: stdlib, sqlx, or sqlc")

	if err := fs.Parse(reorderArgs(args, map[string]bool{
		"--module":        true,
		"--preset":        true,
		"--with":          true,
		"--fiber-version": true,
		"--cli-style":     true,
		"--logger":        true,
		"--db":            true,
		"--data-access":   true,
	})); err != nil {
		return err
	}

	positionals := fs.Args()
	if len(positionals) != 1 {
		return errors.New("new requires exactly one project name")
	}

	projectName := positionals[0]
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	targetDir := filepath.Join(cwd, projectName)
	options := map[string]string{
		"command":                "new",
		"output_mode":            "new",
		"target_dir":             targetDir,
		stack.OptionFiberVersion: *fiberVersion,
		stack.OptionCLIStyle:     *cliStyle,
	}
	setOptionalRuntimeFlags(fs, options, *loggerBackend, *dbKind, *dataAccess)
	req := core.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options:      options,
	}

	summary, err := core.Run(req)
	if err != nil {
		return err
	}

	printSummary(os.Stdout, summary)
	return nil
}

func runInit(args []string) error {
	fs := newFlagSet("init")
	name := fs.String("name", "", "project name override")
	modulePath := fs.String("module", "", "go module path")
	preset := fs.String("preset", "light", "preset name")
	with := fs.String("with", "", "comma-separated capability names")
	fiberVersion := fs.String("fiber-version", stack.DefaultFiberVersion(), "fiber version: v3 or v2")
	cliStyle := fs.String("cli-style", stack.DefaultCLIStyle(), "cli style: cobra or native")
	loggerBackend := fs.String("logger", stack.DefaultLogger(), "logger backend: zap or slog")
	dbKind := fs.String("db", stack.DefaultDB(), "database kind: sqlite, pgsql, or mysql")
	dataAccess := fs.String("data-access", stack.DefaultDataAccess(), "data access stack: stdlib, sqlx, or sqlc")

	if err := fs.Parse(reorderArgs(args, map[string]bool{
		"--name":          true,
		"--module":        true,
		"--preset":        true,
		"--with":          true,
		"--fiber-version": true,
		"--cli-style":     true,
		"--logger":        true,
		"--db":            true,
		"--data-access":   true,
	})); err != nil {
		return err
	}

	if len(fs.Args()) != 0 {
		return errors.New("init does not accept positional arguments")
	}

	projectName := *name
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectName = filepath.Base(cwd)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	options := map[string]string{
		"command":                "init",
		"output_mode":            "init",
		"target_dir":             cwd,
		stack.OptionFiberVersion: *fiberVersion,
		stack.OptionCLIStyle:     *cliStyle,
	}
	setOptionalRuntimeFlags(fs, options, *loggerBackend, *dbKind, *dataAccess)

	req := core.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options:      options,
	}

	summary, err := core.Run(req)
	if err != nil {
		return err
	}

	printSummary(os.Stdout, summary)
	return nil
}

func runList(args []string) error {
	if len(args) != 1 {
		return errors.New("list requires one target: presets or capabilities")
	}

	catalog, err := loadCatalog()
	if err != nil {
		return err
	}

	switch args[0] {
	case "presets":
		for _, preset := range catalog.Presets {
			fmt.Printf("%s\timplemented=%t\t%s\n", preset.Name, preset.Implemented, preset.Summary)
		}
		return nil
	case "capabilities":
		for _, capability := range catalog.Capabilities {
			fmt.Printf("%s\timplemented=%t\t%s\n", capability.Name, capability.Implemented, capability.Summary)
		}
		return nil
	default:
		return fmt.Errorf("unknown list target %q", args[0])
	}
}

func runExplain(args []string) error {
	if len(args) != 2 {
		return errors.New("explain requires a kind and a name")
	}

	catalog, err := loadCatalog()
	if err != nil {
		return err
	}

	switch args[0] {
	case "preset":
		preset, ok := catalog.FindPreset(args[1])
		if !ok {
			return fmt.Errorf("unknown preset %q", args[1])
		}
		fmt.Printf("preset: %s\nsummary: %s\ndescription: %s\nimplemented: %t\nbase: %s\npacks: %s\ndefault_capabilities: %s\nallowed_capabilities: %s\ndefault_stack: %s\ndefault_logger: %s\ndefault_database: %s\ndefault_data_access: %s\nsupported_fiber_versions: %s\nsupported_cli_styles: %s\n", preset.Name, preset.Summary, preset.Description, preset.Implemented, joinOrNone([]string{preset.Base}), joinOrNone(preset.Packs), joinOrNone(preset.DefaultCapabilities), joinOrNone(preset.AllowedCapabilities), stack.DefaultStackLabel(), defaultLoggerForPreset(preset.Name), defaultDatabaseForPreset(preset.Name), defaultDataAccessForPreset(preset.Name), stack.SupportedFiberVersions(), stack.SupportedCLIStyles())
		if preset.Name == "extra-light" {
			fmt.Println("phase11_runtime_options: unsupported")
		} else {
			fmt.Printf("supported_loggers: %s\nsupported_databases: %s\nsupported_data_access: %s\n", stack.SupportedLoggers(), stack.SupportedDatabases(), stack.SupportedDataAccess())
		}
		return nil
	case "capability":
		capability, ok := catalog.FindCapability(args[1])
		if !ok {
			return fmt.Errorf("unknown capability %q", args[1])
		}
		defaultOn, optionalOn, unsupportedOn := capabilityPresetBoundary(catalog, capability)
		fmt.Printf("capability: %s\nsummary: %s\ndescription: %s\nimplemented: %t\npacks: %s\nallowed_presets: %s\ndefault_on_presets: %s\noptional_on_presets: %s\nunsupported_on_presets: %s\ndepends_on: %s\nconflicts_with: %s\n", capability.Name, capability.Summary, capability.Description, capability.Implemented, joinOrNone(capability.Packs), joinOrNone(orderNames(capability.AllowedPresets, []string{"heavy", "medium", "light", "extra-light"})), joinOrNone(defaultOn), joinOrNone(optionalOn), joinOrNone(unsupportedOn), joinOrNone(capability.DependsOn), joinOrNone(capability.ConflictsWith))
		return nil
	default:
		return fmt.Errorf("unknown explain target %q", args[0])
	}
}

func runValidate(args []string) error {
	if len(args) != 0 {
		return errors.New("validate does not accept positional arguments")
	}

	catalog, err := loadCatalog()
	if err != nil {
		return err
	}
	if err := validator.ValidateCatalog(catalog); err != nil {
		return err
	}
	if err := validator.ValidateAssets(manifest.ResolveRoot(""), catalog); err != nil {
		return err
	}

	fmt.Printf("state 4 generator validated successfully: presets=%d capabilities=%d replace_rules=%d injection_rules=%d\n", len(catalog.Presets), len(catalog.Capabilities), len(catalog.ReplaceRules), len(catalog.InjectionRules))
	fmt.Printf("implemented presets: %s\n", joinOrNone(implementedPresets(catalog)))
	fmt.Printf("implemented capabilities: %s\n", joinOrNone(implementedCapabilities(catalog)))
	fmt.Printf("deferred capabilities: %s\n", joinOrNone(deferredCapabilities(catalog)))
	fmt.Println("stable production baseline: medium")
	fmt.Println("completed production track: heavy")
	fmt.Println("current stage: phase-15-build-and-post-generation-engineering")
	fmt.Println("phase 9 delivery: completed")
	fmt.Println("phase 10 delivery: completed")
	fmt.Println("phase 11 delivery: completed")
	fmt.Println("phase 12 delivery: completed")
	fmt.Println("phase 13 delivery: completed")
	fmt.Println("phase 14 delivery: completed")
	fmt.Println("phase 15 focus: build and post-generation engineering")
	fmt.Println("phase 15 delivery target: fiberx build and release-oriented output management")
	fmt.Println("default medium experience: swagger,embedded-ui")
	fmt.Println("default heavy experience: swagger,embedded-ui")
	fmt.Println("light optional experience: swagger,embedded-ui")
	fmt.Println("extra-light optional experience: none")
	printCapabilityPolicy(os.Stdout, catalog)
	fmt.Printf("default stack: %s\n", stack.DefaultStackLabel())
	fmt.Printf("supported fiber versions: %s\n", stack.SupportedFiberVersions())
	fmt.Printf("supported cli styles: %s\n", stack.SupportedCLIStyles())
	fmt.Printf("default logger: %s\n", stack.DefaultLogger())
	fmt.Printf("default database: %s\n", stack.DefaultDB())
	fmt.Printf("default data access: %s\n", stack.DefaultDataAccess())
	fmt.Printf("supported loggers: %s\n", stack.SupportedLoggers())
	fmt.Printf("supported databases: %s\n", stack.SupportedDatabases())
	fmt.Printf("supported data access: %s\n", stack.SupportedDataAccess())
	fmt.Println("phase 11 first-round presets: medium,heavy,light")
	fmt.Println("phase 11 deferred presets: extra-light")
	return nil
}

func runDoctor(args []string) error {
	if len(args) != 0 {
		return errors.New("doctor does not accept positional arguments")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	root := manifest.ResolveRoot("")
	catalog, err := manifest.LoadCatalog(root)
	if err != nil {
		return err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	fmt.Printf("cwd: %s\n", cwd)
	fmt.Printf("go: %s\n", runtime.Version())
	fmt.Printf("state: %s\n", "state-4")
	fmt.Printf("phase: %s\n", "phase-15-build-and-post-generation-engineering")
	fmt.Printf("manifest-root: %s\n", rootAbs)
	fmt.Printf("presets: %d\n", len(catalog.Presets))
	fmt.Printf("capabilities: %d\n", len(catalog.Capabilities))
	fmt.Printf("replace-rules: %d\n", len(catalog.ReplaceRules))
	fmt.Printf("injection-rules: %d\n", len(catalog.InjectionRules))
	fmt.Printf("implemented-presets: %s\n", joinOrNone(implementedPresets(catalog)))
	fmt.Printf("implemented-capabilities: %s\n", joinOrNone(implementedCapabilities(catalog)))
	fmt.Printf("deferred-capabilities: %s\n", joinOrNone(deferredCapabilities(catalog)))
	fmt.Printf("medium-production-baseline: %s\n", "stable")
	fmt.Printf("heavy-production-track: %s\n", "completed")
	fmt.Printf("phase-9-stack-normalization: %s\n", "completed")
	fmt.Printf("phase-10-capability-consolidation: %s\n", "completed")
	fmt.Printf("phase-11-runtime-options-and-data-access: %s\n", "completed")
	fmt.Printf("phase-12-capability-level-verification: %s\n", "completed")
	fmt.Printf("phase-13-version-upgrade-and-diff-detection: %s\n", "completed")
	fmt.Printf("phase-14-upgrade-assistant-and-compatibility-policy: %s\n", "completed")
	fmt.Printf("phase-15-build-and-post-generation-engineering: %s\n", "active")
	fmt.Printf("phase-15-focus: %s\n", "build-and-post-generation-engineering")
	fmt.Printf("phase-15-delivery-target: %s\n", "fiberx-build-and-release-oriented-output-management")
	fmt.Printf("default-medium-capabilities: %s\n", "swagger,embedded-ui")
	fmt.Printf("default-heavy-capabilities: %s\n", "swagger,embedded-ui")
	fmt.Printf("light-optional-capabilities: %s\n", "swagger,embedded-ui")
	fmt.Printf("extra-light-optional-capabilities: %s\n", "none")
	printCapabilityPolicy(os.Stdout, catalog)
	fmt.Printf("default-stack: %s\n", stack.DefaultStackLabel())
	fmt.Printf("supported-fiber-versions: %s\n", stack.SupportedFiberVersions())
	fmt.Printf("supported-cli-styles: %s\n", stack.SupportedCLIStyles())
	fmt.Printf("default-logger: %s\n", stack.DefaultLogger())
	fmt.Printf("default-database: %s\n", stack.DefaultDB())
	fmt.Printf("default-data-access: %s\n", stack.DefaultDataAccess())
	fmt.Printf("supported-loggers: %s\n", stack.SupportedLoggers())
	fmt.Printf("supported-databases: %s\n", stack.SupportedDatabases())
	fmt.Printf("supported-data-access: %s\n", stack.SupportedDataAccess())
	fmt.Printf("phase-11-first-round-presets: %s\n", "medium,heavy,light")
	fmt.Printf("phase-11-deferred-presets: %s\n", "extra-light")
	fmt.Printf("generator-version: %s\n", version.Version)
	fmt.Printf("generator-commit: %s\n", version.Commit)
	fmt.Println("writer-mode: real-write")
	return nil
}

func runInspect(args []string) error {
	fs := newFlagSet("inspect")
	asJSON := fs.Bool("json", false, "render inspect output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}
	projectManifest, err := metadata.LoadManifest(projectDir)
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, projectManifest)
	}

	fmt.Printf("project: %s\n", projectDir)
	fmt.Printf("metadata: %s\n", filepath.ToSlash(filepath.Join(metadata.ManifestDir, metadata.ManifestFilename)))
	fmt.Printf("generated-at: %s\n", projectManifest.GeneratedAt)
	fmt.Printf("generator-version: %s\n", projectManifest.Generator.Version)
	fmt.Printf("generator-commit: %s\n", projectManifest.Generator.Commit)
	fmt.Printf("preset: %s\n", projectManifest.Recipe.Preset)
	fmt.Printf("capabilities: %s\n", joinOrNone(projectManifest.Recipe.Capabilities))
	fmt.Printf("stack: fiber-%s + %s\n", projectManifest.Recipe.FiberVersion, projectManifest.Recipe.CLIStyle)
	if projectManifest.Recipe.Logger != "" || projectManifest.Recipe.DB != "" || projectManifest.Recipe.DataAccess != "" {
		fmt.Printf("runtime: logger=%s db=%s data-access=%s\n", valueOrNone(projectManifest.Recipe.Logger), valueOrNone(projectManifest.Recipe.DB), valueOrNone(projectManifest.Recipe.DataAccess))
	}
	fmt.Printf("base: %s\n", projectManifest.Assets.Base)
	fmt.Printf("preset packs: %s\n", joinOrNone(projectManifest.Assets.PresetPacks))
	fmt.Printf("capability packs: %s\n", joinOrNone(projectManifest.Assets.CapabilityPacks))
	fmt.Printf("runtime overlays: %s\n", joinOrNone(projectManifest.Assets.RuntimeOverlays))
	fmt.Printf("replace rules: %s\n", joinOrNone(projectManifest.Assets.ReplaceRules))
	fmt.Printf("injection rules: %s\n", joinOrNone(projectManifest.Assets.InjectionRules))
	fmt.Printf("template fingerprint: %s\n", projectManifest.Fingerprints.TemplateSet)
	fmt.Printf("rendered fingerprint: %s\n", projectManifest.Fingerprints.RenderedOutput)
	fmt.Printf("managed files: %d\n", len(projectManifest.ManagedFiles))
	return nil
}

func runDiff(args []string) error {
	fs := newFlagSet("diff")
	asJSON := fs.Bool("json", false, "render diff output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}

	diffReport, err := metadata.BuildDiff(projectDir, manifest.ResolveRoot(""))
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, diffReport)
	}

	fmt.Printf("project: %s\n", projectDir)
	fmt.Printf("status: %s\n", diffReport.Status)
	fmt.Printf("generated-by: %s (%s)\n", diffReport.Generator.Generated.Version, diffReport.Generator.Generated.Commit)
	fmt.Printf("current-generator: %s (%s)\n", diffReport.Generator.Current.Version, diffReport.Generator.Current.Commit)
	fmt.Printf("preset: %s\n", diffReport.Recipe.Preset)
	fmt.Printf("capabilities: %s\n", joinOrNone(diffReport.Recipe.Capabilities))
	fmt.Printf("missing files: %s\n", joinOrNone(diffReport.MissingFiles))
	fmt.Printf("changed files: %s\n", joinOrNone(diffReport.ChangedFiles))
	fmt.Printf("new managed files: %s\n", joinOrNone(diffReport.NewManagedFiles))
	fmt.Printf("generator drift files: %s\n", joinOrNone(diffReport.GeneratorDriftFiles))
	return nil
}

func runUpgrade(args []string) error {
	if len(args) == 0 {
		return errors.New("upgrade requires a subcommand: inspect or plan")
	}

	switch args[0] {
	case "inspect":
		return runUpgradeInspect(args[1:])
	case "plan":
		return runUpgradePlan(args[1:])
	default:
		return fmt.Errorf("unknown upgrade subcommand %q", args[0])
	}
}

func runUpgradeInspect(args []string) error {
	fs := newFlagSet("upgrade inspect")
	asJSON := fs.Bool("json", false, "render upgrade inspect output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}

	assessment, err := upgrade.Inspect(projectDir, manifest.ResolveRoot(""))
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, assessment)
	}

	fmt.Printf("project: %s\n", assessment.ProjectDir)
	fmt.Printf("generated-by: %s (%s)\n", assessment.GeneratedGenerator.Version, assessment.GeneratedGenerator.Commit)
	fmt.Printf("current-generator: %s (%s)\n", assessment.CurrentGenerator.Version, assessment.CurrentGenerator.Commit)
	fmt.Printf("preset: %s\n", assessment.Recipe.Preset)
	fmt.Printf("capabilities: %s\n", joinOrNone(assessment.Recipe.Capabilities))
	fmt.Printf("stack: fiber-%s + %s\n", assessment.Recipe.FiberVersion, assessment.Recipe.CLIStyle)
	if assessment.Recipe.Logger != "" || assessment.Recipe.DB != "" || assessment.Recipe.DataAccess != "" {
		fmt.Printf("runtime: logger=%s db=%s data-access=%s\n", valueOrNone(assessment.Recipe.Logger), valueOrNone(assessment.Recipe.DB), valueOrNone(assessment.Recipe.DataAccess))
	}
	fmt.Printf("diff status: %s\n", assessment.DiffStatus)
	fmt.Printf("compatibility level: %s\n", assessment.CompatibilityLevel)
	fmt.Printf("reasons: %s\n", joinOrNone(assessment.Reasons))
	fmt.Printf("blocking issues: %s\n", joinOrNone(assessment.BlockingIssues))
	fmt.Printf("local modified files: %s\n", joinOrNone(assessment.LocalModifiedFiles))
	fmt.Printf("generator drift files: %s\n", joinOrNone(assessment.GeneratorDriftFiles))
	fmt.Printf("missing files: %s\n", joinOrNone(assessment.MissingFiles))
	fmt.Printf("new managed files: %s\n", joinOrNone(assessment.NewManagedFiles))
	return nil
}

func runUpgradePlan(args []string) error {
	fs := newFlagSet("upgrade plan")
	asJSON := fs.Bool("json", false, "render upgrade plan output as JSON")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--json": true})); err != nil {
		return err
	}

	projectDir, err := resolveProjectDir(fs.Args())
	if err != nil {
		return err
	}

	upgradePlan, err := upgrade.Plan(projectDir, manifest.ResolveRoot(""))
	if err != nil {
		return err
	}

	if *asJSON {
		return writeJSON(os.Stdout, upgradePlan)
	}

	fmt.Printf("project: %s\n", upgradePlan.Assessment.ProjectDir)
	fmt.Printf("compatibility level: %s\n", upgradePlan.Assessment.CompatibilityLevel)
	fmt.Printf("diff status: %s\n", upgradePlan.Assessment.DiffStatus)
	fmt.Printf("upgrade summary: %s\n", joinOrNone(upgradePlan.Assessment.Reasons))
	fmt.Printf("blocking issues: %s\n", joinOrNone(upgradePlan.Assessment.BlockingIssues))
	fmt.Printf("local modified files: %s\n", joinOrNone(upgradePlan.Assessment.LocalModifiedFiles))
	fmt.Printf("generator drift files: %s\n", joinOrNone(upgradePlan.Assessment.GeneratorDriftFiles))
	fmt.Printf("missing files: %s\n", joinOrNone(upgradePlan.Assessment.MissingFiles))
	fmt.Printf("new managed files: %s\n", joinOrNone(upgradePlan.Assessment.NewManagedFiles))
	fmt.Printf("managed files to review: %s\n", joinOrNone(upgradePlan.ManagedFilesToReview))
	if len(upgradePlan.RecommendedSteps) == 0 {
		fmt.Println("recommended steps: (none)")
		return nil
	}
	fmt.Println("recommended steps:")
	for index, step := range upgradePlan.RecommendedSteps {
		fmt.Printf("  %d. %s\n", index+1, step)
	}
	return nil
}

func runBuild(args []string) error {
	fs := newFlagSet("build")
	clean := fs.Bool("clean", false, "clean the output directory before building")
	platform := fs.String("target", "", "filter builds to a single goos/goarch platform")
	if err := fs.Parse(reorderArgs(args, map[string]bool{"--target": true})); err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := buildconfig.Load(projectDir)
	if err != nil {
		return err
	}

	result, err := build.Execute(projectDir, cfg, build.Options{
		TargetNames:    fs.Args(),
		PlatformFilter: *platform,
		Clean:          *clean,
	})
	if err != nil {
		return err
	}

	fmt.Printf("built project=%s out_dir=%s\n", cfg.Project.Name, filepath.ToSlash(result.OutDir))
	fmt.Printf("version: %s\n", result.Version.Version)
	fmt.Printf("commit: %s\n", result.Version.Commit)
	fmt.Printf("build time: %s\n", result.Version.BuildTime)
	fmt.Printf("artifacts: %d\n", len(result.Artifacts))
	for _, artifact := range result.Artifacts {
		fmt.Printf("  - target=%s platform=%s output=%s\n", artifact.TargetName, artifact.Platform, filepath.ToSlash(artifact.OutputPath))
	}
	return nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "fiberx is the State 4 generator CLI with a stable medium baseline, a completed heavy production track, and active build and post-generation engineering work on top of fiber-v3 + cobra + viper defaults.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  fiberx new <name> [--module path] [--preset name] [--with cap1,cap2] [--fiber-version v3|v2] [--cli-style cobra|native] [--logger zap|slog] [--db sqlite|pgsql|mysql] [--data-access stdlib|sqlx|sqlc]")
	fmt.Fprintln(w, "  fiberx init [--name name] [--module path] [--preset name] [--with cap1,cap2] [--fiber-version v3|v2] [--cli-style cobra|native] [--logger zap|slog] [--db sqlite|pgsql|mysql] [--data-access stdlib|sqlx|sqlc]")
	fmt.Fprintln(w, "  fiberx list presets")
	fmt.Fprintln(w, "  fiberx list capabilities")
	fmt.Fprintln(w, "  fiberx explain preset <name>")
	fmt.Fprintln(w, "  fiberx explain capability <name>")
	fmt.Fprintln(w, "  fiberx inspect [path] [--json]")
	fmt.Fprintln(w, "  fiberx diff [path] [--json]")
	fmt.Fprintln(w, "  fiberx upgrade inspect [path] [--json]")
	fmt.Fprintln(w, "  fiberx upgrade plan [path] [--json]")
	fmt.Fprintln(w, "  fiberx build [target...] [--clean] [--target goos/goarch]")
	fmt.Fprintln(w, "  fiberx validate")
	fmt.Fprintln(w, "  fiberx doctor")
	fmt.Fprintf(w, "\nDefault stack: %s\n", stack.DefaultStackLabel())
	fmt.Fprintf(w, "Default logger/database/data access: %s / %s / %s\n", stack.DefaultLogger(), stack.DefaultDB(), stack.DefaultDataAccess())
	fmt.Fprintln(w, "Capability policy: swagger and embedded-ui default on medium/heavy, optional on light; redis optional on medium/heavy only.")
	fmt.Fprintln(w, "Phase 11 runtime policy: medium/heavy/light support logger/db/data-access selection; extra-light rejects these options.")
	fmt.Fprintln(w, "Current roadmap stage: Phase 15 build and post-generation engineering.")
	fmt.Fprintln(w, "Phase 15 focus: build and release-oriented post-generation engineering.")
	fmt.Fprintln(w, "Phase 15 delivery target: fiberx build and release-oriented output management.")
}

func loadCatalog() (manifest.Catalog, error) {
	return manifest.LoadCatalog(manifest.ResolveRoot(""))
}

func implementedPresets(catalog manifest.Catalog) []string {
	names := make([]string, 0, len(catalog.Presets))
	for _, preset := range catalog.Presets {
		if preset.Implemented {
			names = append(names, preset.Name)
		}
	}
	return orderNames(names, []string{"heavy", "medium", "light", "extra-light"})
}

func implementedCapabilities(catalog manifest.Catalog) []string {
	names := make([]string, 0, len(catalog.Capabilities))
	for _, capability := range catalog.Capabilities {
		if capability.Implemented {
			names = append(names, capability.Name)
		}
	}
	return orderNames(names, []string{"redis", "swagger", "embedded-ui"})
}

func deferredCapabilities(catalog manifest.Catalog) []string {
	names := make([]string, 0, len(catalog.Capabilities))
	for _, capability := range catalog.Capabilities {
		if capability.Implemented {
			continue
		}
		names = append(names, capability.Name)
	}
	return orderNames(names, []string{"swagger", "embedded-ui", "redis"})
}

func joinOrNone(items []string) string {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		return "(none)"
	}
	return strings.Join(filtered, ",")
}

func orderNames(items []string, preferred []string) []string {
	if len(items) <= 1 {
		return append([]string(nil), items...)
	}

	order := make(map[string]int, len(preferred))
	for index, name := range preferred {
		order[name] = index
	}

	ordered := append([]string(nil), items...)
	sort.SliceStable(ordered, func(i int, j int) bool {
		left, leftOK := order[ordered[i]]
		right, rightOK := order[ordered[j]]
		switch {
		case leftOK && rightOK:
			return left < right
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return ordered[i] < ordered[j]
		}
	})

	return ordered
}

func capabilityPresetBoundary(catalog manifest.Catalog, capability manifest.CapabilityManifest) ([]string, []string, []string) {
	defaultOn := []string{}
	optionalOn := []string{}
	unsupportedOn := []string{}
	for _, presetName := range implementedPresets(catalog) {
		preset, ok := catalog.FindPreset(presetName)
		if !ok {
			continue
		}
		if contains(preset.DefaultCapabilities, capability.Name) {
			defaultOn = append(defaultOn, presetName)
			continue
		}
		if contains(capability.AllowedPresets, presetName) {
			optionalOn = append(optionalOn, presetName)
			continue
		}
		unsupportedOn = append(unsupportedOn, presetName)
	}
	return defaultOn, optionalOn, unsupportedOn
}

func printCapabilityPolicy(w io.Writer, catalog manifest.Catalog) {
	for _, capabilityName := range implementedCapabilities(catalog) {
		capability, ok := catalog.FindCapability(capabilityName)
		if !ok {
			continue
		}
		defaultOn, optionalOn, unsupportedOn := capabilityPresetBoundary(catalog, capability)
		fmt.Fprintf(w, "capability-policy-%s: default=%s optional=%s unsupported=%s\n", capability.Name, joinOrNone(defaultOn), joinOrNone(optionalOn), joinOrNone(unsupportedOn))
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func printSummary(w io.Writer, summary report.Summary) {
	fmt.Fprintf(w, "generated preset=%s target=%s\n", summary.Preset, summary.TargetDir)
	fmt.Fprintf(w, "stack: fiber-%s + %s", summary.FiberVersion, summary.CLIStyle)
	if summary.CLIStyle == stack.CLICobra {
		fmt.Fprintf(w, " + viper")
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "runtime: logger=%s db=%s data-access=%s\n", summary.Logger, summary.Database, summary.DataAccess)
	fmt.Fprintf(w, "base: %s\n", summary.Base)
	fmt.Fprintf(w, "preset packs: %s\n", joinOrNone(summary.PresetPacks))
	fmt.Fprintf(w, "capabilities: %s\n", joinOrNone(summary.Capabilities))
	fmt.Fprintf(w, "capability packs: %s\n", joinOrNone(summary.CapabilityPacks))
	fmt.Fprintf(w, "runtime overlays: %s\n", joinOrNone(summary.RuntimeOverlays))
	fmt.Fprintf(w, "replace rules: %s\n", joinOrNone(summary.ReplaceRules))
	fmt.Fprintf(w, "injection rules: %s\n", joinOrNone(summary.InjectionRules))
	fmt.Fprintf(w, "generator: %s (%s)\n", summary.GeneratorVersion, summary.GeneratorCommit)
	fmt.Fprintf(w, "template fingerprint: %s\n", summary.TemplateSetFingerprint)
	fmt.Fprintf(w, "rendered fingerprint: %s\n", summary.RenderedOutputFingerprint)
	fmt.Fprintf(w, "metadata: %s\n", summary.MetadataPath)
	fmt.Fprintf(w, "written files: %d\n", summary.WrittenFiles)
	for _, path := range summary.WrittenPaths {
		fmt.Fprintf(w, "  - %s\n", path)
	}
	if len(summary.Warnings) > 0 {
		fmt.Fprintf(w, "warnings: %s\n", joinOrNone(summary.Warnings))
	}
}

func defaultLoggerForPreset(presetName string) string {
	if presetName == "extra-light" {
		return "slog"
	}
	return stack.DefaultLogger()
}

func defaultDatabaseForPreset(presetName string) string {
	return stack.DefaultDB()
}

func defaultDataAccessForPreset(presetName string) string {
	if presetName == "extra-light" {
		return "builtin"
	}
	return stack.DefaultDataAccess()
}

func resolveProjectDir(args []string) (string, error) {
	if len(args) > 1 {
		return "", errors.New("command accepts at most one project path")
	}
	if len(args) == 0 {
		return os.Getwd()
	}
	return filepath.Abs(args[0])
}

func writeJSON(w io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func valueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(none)"
	}
	return value
}

func parseCapabilities(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	capabilities := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		capabilities = append(capabilities, name)
	}

	return capabilities
}

func defaultModulePath(projectName string, explicit string) string {
	if explicit != "" {
		return explicit
	}

	slug := strings.ToLower(strings.TrimSpace(projectName))
	slug = strings.ReplaceAll(slug, " ", "-")
	if slug == "" {
		slug = "fiberx-app"
	}

	return "github.com/example/" + slug
}

func reorderArgs(args []string, valueFlags map[string]bool) []string {
	reordered := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))

	for index := 0; index < len(args); index++ {
		current := args[index]
		if strings.HasPrefix(current, "-") {
			reordered = append(reordered, current)
			if valueFlags[current] && index+1 < len(args) {
				index++
				reordered = append(reordered, args[index])
			}
			continue
		}

		positionals = append(positionals, current)
	}

	return append(reordered, positionals...)
}

func setOptionalRuntimeFlags(fs *flag.FlagSet, options map[string]string, loggerBackend, dbKind, dataAccess string) {
	visited := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	if visited["logger"] {
		options[stack.OptionLogger] = loggerBackend
	}
	if visited["db"] {
		options[stack.OptionDB] = dbKind
	}
	if visited["data-access"] {
		options[stack.OptionDataAccess] = dataAccess
	}
}
