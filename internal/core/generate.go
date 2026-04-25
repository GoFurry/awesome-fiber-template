package core

import (
	"github.com/GoFurry/fiberx/internal/manifest"
	"github.com/GoFurry/fiberx/internal/planner"
	"github.com/GoFurry/fiberx/internal/renderer"
	"github.com/GoFurry/fiberx/internal/report"
	"github.com/GoFurry/fiberx/internal/validator"
	"github.com/GoFurry/fiberx/internal/writer"
)

func Generate(req Request) error {
	catalogRoot := manifest.ResolveRoot(req.Options["manifest_root"])

	catalog, err := manifest.LoadCatalog(catalogRoot)
	if err != nil {
		return err
	}

	if err := validator.ValidateCatalog(catalog); err != nil {
		return err
	}
	if err := validator.ValidateAssets(catalogRoot, catalog); err != nil {
		return err
	}

	if err := validator.ValidateRequest(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, catalog); err != nil {
		return err
	}

	preset, _ := catalog.FindPreset(req.Preset)
	selectedCapabilities := make([]manifest.CapabilityManifest, 0, len(req.Capabilities))
	for _, name := range req.Capabilities {
		capability, ok := catalog.FindCapability(name)
		if !ok {
			continue
		}
		selectedCapabilities = append(selectedCapabilities, capability)
	}
	if err := validator.ValidateGenerationSupport(preset, selectedCapabilities); err != nil {
		return err
	}

	plan := planner.BuildPlan(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, req.Options, catalogRoot, catalog)
	rendered, err := renderer.Render(plan)
	if err != nil {
		return err
	}
	writeResult, err := writer.New(req.Options["target_dir"]).Write(rendered)
	if err != nil {
		return err
	}

	_ = report.Build(plan, rendered, writeResult)

	return nil
}
