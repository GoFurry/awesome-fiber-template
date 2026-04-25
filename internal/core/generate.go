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
	catalogRoot := manifest.DefaultRoot()
	if root := req.Options["manifest_root"]; root != "" {
		catalogRoot = root
	}

	catalog, err := manifest.LoadCatalog(catalogRoot)
	if err != nil {
		return err
	}

	if err := validator.ValidateCatalog(catalog); err != nil {
		return err
	}

	if err := validator.ValidateRequest(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, catalog); err != nil {
		return err
	}

	plan := planner.BuildPlan(req.ProjectName, req.ModulePath, req.Preset, req.Capabilities, req.Options, catalog)
	rendered := renderer.Render(plan)
	writeResult, err := writer.NewDryRunWriter().Write(rendered)
	if err != nil {
		return err
	}

	_ = report.Build(plan, rendered, writeResult)

	return nil
}
