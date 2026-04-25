package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/GoFurry/fiberx"
	"github.com/GoFurry/fiberx/internal/manifest"
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
	req := fiberx.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options: map[string]string{
			"command": "new",
		},
	}

	if err := fiberx.Generate(req); err != nil {
		return err
	}

	fmt.Printf("Phase 3 dry-run accepted request: project=%q preset=%q module=%q capabilities=%d\n", req.ProjectName, req.Preset, req.ModulePath, len(req.Capabilities))
	fmt.Println("Declaration loading, validation, and planning ran successfully. Project file writing stays deferred to later phases.")
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

	req := fiberx.Request{
		ProjectName:  projectName,
		ModulePath:   defaultModulePath(projectName, *modulePath),
		Preset:       *preset,
		Capabilities: parseCapabilities(*with),
		Options: map[string]string{
			"command": "init",
		},
	}

	if err := fiberx.Generate(req); err != nil {
		return err
	}

	fmt.Printf("Phase 3 dry-run initialized request for current directory: project=%q preset=%q module=%q capabilities=%d\n", req.ProjectName, req.Preset, req.ModulePath, len(req.Capabilities))
	fmt.Println("Declaration loading, validation, and planning ran successfully. Asset rendering and writing stay deferred to later phases.")
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
			fmt.Printf("%s\t%s\n", preset.Name, preset.Summary)
		}
		return nil
	case "capabilities":
		for _, capability := range catalog.Capabilities {
			fmt.Printf("%s\t%s\n", capability.Name, capability.Summary)
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
		fmt.Printf("preset: %s\nsummary: %s\ndescription: %s\nbase: %s\ndefault_capabilities: %s\nallowed_capabilities: %s\n", preset.Name, preset.Summary, preset.Description, preset.Base, strings.Join(preset.DefaultCapabilities, ","), strings.Join(preset.AllowedCapabilities, ","))
		return nil
	case "capability":
		capability, ok := catalog.FindCapability(args[1])
		if !ok {
			return fmt.Errorf("unknown capability %q", args[1])
		}
		fmt.Printf("capability: %s\nsummary: %s\ndescription: %s\nallowed_presets: %s\ndepends_on: %s\nconflicts_with: %s\n", capability.Name, capability.Summary, capability.Description, strings.Join(capability.AllowedPresets, ","), strings.Join(capability.DependsOn, ","), strings.Join(capability.ConflictsWith, ","))
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

	fmt.Printf("Phase 3 declarations validated successfully: presets=%d capabilities=%d replace_rules=%d injection_rules=%d\n", len(catalog.Presets), len(catalog.Capabilities), len(catalog.ReplaceRules), len(catalog.InjectionRules))
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

	root := manifest.DefaultRoot()
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
	fmt.Printf("phase: %s\n", "phase-3-declarations")
	fmt.Printf("manifest-root: %s\n", rootAbs)
	fmt.Printf("presets: %d\n", len(catalog.Presets))
	fmt.Printf("capabilities: %d\n", len(catalog.Capabilities))
	fmt.Printf("replace-rules: %d\n", len(catalog.ReplaceRules))
	fmt.Printf("injection-rules: %d\n", len(catalog.InjectionRules))
	fmt.Println("writer-mode: dry-run")
	return nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "fiberx is the Phase 3 generator CLI with disk-backed declarations.")
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
	return manifest.LoadCatalog(manifest.DefaultRoot())
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
