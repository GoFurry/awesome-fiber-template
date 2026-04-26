package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/GoFurry/fiberx/internal/core"
	"github.com/GoFurry/fiberx/internal/manifest"
	"github.com/GoFurry/fiberx/internal/report"
	"github.com/GoFurry/fiberx/internal/validator"
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

	if err := fs.Parse(reorderArgs(args, map[string]bool{
		"--module": true,
		"--preset": true,
		"--with":   true,
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
	req := core.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options: map[string]string{
			"command":     "new",
			"output_mode": "new",
			"target_dir":  targetDir,
		},
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

	if err := fs.Parse(reorderArgs(args, map[string]bool{
		"--name":   true,
		"--module": true,
		"--preset": true,
		"--with":   true,
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

	req := core.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options: map[string]string{
			"command":     "init",
			"output_mode": "init",
			"target_dir":  cwd,
		},
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
		fmt.Printf("preset: %s\nsummary: %s\ndescription: %s\nimplemented: %t\nbase: %s\npacks: %s\ndefault_capabilities: %s\nallowed_capabilities: %s\n", preset.Name, preset.Summary, preset.Description, preset.Implemented, joinOrNone([]string{preset.Base}), joinOrNone(preset.Packs), joinOrNone(preset.DefaultCapabilities), joinOrNone(preset.AllowedCapabilities))
		return nil
	case "capability":
		capability, ok := catalog.FindCapability(args[1])
		if !ok {
			return fmt.Errorf("unknown capability %q", args[1])
		}
		fmt.Printf("capability: %s\nsummary: %s\ndescription: %s\nimplemented: %t\npacks: %s\nallowed_presets: %s\ndepends_on: %s\nconflicts_with: %s\n", capability.Name, capability.Summary, capability.Description, capability.Implemented, joinOrNone(capability.Packs), joinOrNone(capability.AllowedPresets), joinOrNone(capability.DependsOn), joinOrNone(capability.ConflictsWith))
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

	fmt.Printf("state 1 generator validated successfully: presets=%d capabilities=%d replace_rules=%d injection_rules=%d\n", len(catalog.Presets), len(catalog.Capabilities), len(catalog.ReplaceRules), len(catalog.InjectionRules))
	fmt.Printf("implemented presets: %s\n", joinOrNone(implementedPresets(catalog)))
	fmt.Printf("implemented capabilities: %s\n", joinOrNone(implementedCapabilities(catalog)))
	fmt.Printf("deferred capabilities: %s\n", joinOrNone(deferredCapabilities(catalog)))
	fmt.Println("default medium experience: swagger,embedded-ui")
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
	fmt.Printf("state: %s\n", "state-1")
	fmt.Printf("phase: %s\n", "phase-6-medium-production-baseline")
	fmt.Printf("manifest-root: %s\n", rootAbs)
	fmt.Printf("presets: %d\n", len(catalog.Presets))
	fmt.Printf("capabilities: %d\n", len(catalog.Capabilities))
	fmt.Printf("replace-rules: %d\n", len(catalog.ReplaceRules))
	fmt.Printf("injection-rules: %d\n", len(catalog.InjectionRules))
	fmt.Printf("implemented-presets: %s\n", joinOrNone(implementedPresets(catalog)))
	fmt.Printf("implemented-capabilities: %s\n", joinOrNone(implementedCapabilities(catalog)))
	fmt.Printf("deferred-capabilities: %s\n", joinOrNone(deferredCapabilities(catalog)))
	fmt.Printf("medium-production-baseline: %s\n", "enabled")
	fmt.Printf("default-medium-capabilities: %s\n", "swagger,embedded-ui")
	fmt.Println("writer-mode: real-write")
	return nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "fiberx is the State 1 generator CLI with a medium production baseline and real project generation.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  fiberx new <name> [--module path] [--preset name] [--with cap1,cap2]")
	fmt.Fprintln(w, "  fiberx init [--name name] [--module path] [--preset name] [--with cap1,cap2]")
	fmt.Fprintln(w, "  fiberx list presets")
	fmt.Fprintln(w, "  fiberx list capabilities")
	fmt.Fprintln(w, "  fiberx explain preset <name>")
	fmt.Fprintln(w, "  fiberx explain capability <name>")
	fmt.Fprintln(w, "  fiberx validate")
	fmt.Fprintln(w, "  fiberx doctor")
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

func printSummary(w io.Writer, summary report.Summary) {
	fmt.Fprintf(w, "generated preset=%s target=%s\n", summary.Preset, summary.TargetDir)
	fmt.Fprintf(w, "base: %s\n", summary.Base)
	fmt.Fprintf(w, "preset packs: %s\n", joinOrNone(summary.PresetPacks))
	fmt.Fprintf(w, "capabilities: %s\n", joinOrNone(summary.Capabilities))
	fmt.Fprintf(w, "capability packs: %s\n", joinOrNone(summary.CapabilityPacks))
	fmt.Fprintf(w, "replace rules: %s\n", joinOrNone(summary.ReplaceRules))
	fmt.Fprintf(w, "injection rules: %s\n", joinOrNone(summary.InjectionRules))
	fmt.Fprintf(w, "written files: %d\n", summary.WrittenFiles)
	for _, path := range summary.WrittenPaths {
		fmt.Fprintf(w, "  - %s\n", path)
	}
	if len(summary.Warnings) > 0 {
		fmt.Fprintf(w, "warnings: %s\n", joinOrNone(summary.Warnings))
	}
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
