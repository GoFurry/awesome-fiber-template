package report

import (
	"github.com/GoFurry/fiberx/internal/planner"
	"github.com/GoFurry/fiberx/internal/renderer"
	"github.com/GoFurry/fiberx/internal/writer"
)

type Summary struct {
	Preset         string
	Capabilities   []string
	ReplaceRules   []string
	InjectionRules []string
	PreviewFiles   []string
	Warnings       []string
	DryRun         bool
}

func Build(plan planner.Plan, rendered renderer.Result, writeResult writer.Result) Summary {
	capabilities := make([]string, 0, len(plan.Capabilities))
	for _, capability := range plan.Capabilities {
		capabilities = append(capabilities, capability.Name)
	}

	return Summary{
		Preset:         plan.Preset.Name,
		Capabilities:   capabilities,
		ReplaceRules:   append([]string(nil), rendered.ReplaceRuleHits...),
		InjectionRules: append([]string(nil), rendered.InjectionHits...),
		PreviewFiles:   append([]string(nil), rendered.PreviewFiles...),
		Warnings:       append([]string(nil), rendered.Warnings...),
		DryRun:         writeResult.DryRun,
	}
}
