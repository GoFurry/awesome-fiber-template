package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/GoFurry/fiberx/internal/manifest"
	"github.com/GoFurry/fiberx/internal/planner"
	"github.com/GoFurry/fiberx/internal/renderer"
	"github.com/GoFurry/fiberx/internal/report"
	"github.com/GoFurry/fiberx/internal/validator"
	"github.com/GoFurry/fiberx/internal/writer"
)

func Generate(req Request) error {
	_, err := Run(req)
	return err
}

func Run(req Request) (report.Summary, error) {
	catalogRoot := manifest.ResolveRoot(req.Options["manifest_root"])

	catalog, err := manifest.LoadCatalog(catalogRoot)
	if err != nil {
		return report.Summary{}, err
	}

	if err := validator.ValidateCatalog(catalog); err != nil {
		return report.Summary{}, err
	}
	if err := validator.ValidateAssets(catalogRoot, catalog); err != nil {
		return report.Summary{}, err
	}

	if err := validator.ValidateRequest(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, catalog); err != nil {
		return report.Summary{}, err
	}

	preset, _ := catalog.FindPreset(req.Preset)
	selectedCapabilityNames := append([]string{}, catalog.AppliedDefaultCapabilities(preset)...)
	selectedCapabilityNames = append(selectedCapabilityNames, req.Capabilities...)
	selectedCapabilities := make([]manifest.CapabilityManifest, 0, len(selectedCapabilityNames))
	for _, name := range selectedCapabilityNames {
		capability, ok := catalog.FindCapability(name)
		if !ok {
			continue
		}
		alreadySelected := false
		for _, selected := range selectedCapabilities {
			if selected.Name == capability.Name {
				alreadySelected = true
				break
			}
		}
		if alreadySelected {
			continue
		}
		selectedCapabilities = append(selectedCapabilities, capability)
	}
	if err := validator.ValidateGenerationSupport(preset, selectedCapabilities); err != nil {
		return report.Summary{}, err
	}

	plan := planner.BuildPlan(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, req.Options, catalogRoot, catalog)
	rendered, err := renderer.Render(plan)
	if err != nil {
		return report.Summary{}, err
	}
	writeResult, err := writer.New(req.Options["target_dir"]).Write(rendered)
	if err != nil {
		return report.Summary{}, err
	}
	if err := finalizeGeneratedModule(writeResult.TargetDir); err != nil {
		return report.Summary{}, err
	}

	return report.Build(plan, rendered, writeResult), nil
}

func finalizeGeneratedModule(targetDir string) error {
	if targetDir == "" {
		return nil
	}

	goModPath := filepath.Join(targetDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat generated go.mod: %w", err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("finalize generated module with go mod tidy: %w\n%s", err, string(output))
	}

	return nil
}
