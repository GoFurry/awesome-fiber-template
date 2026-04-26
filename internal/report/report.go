package report

import (
	"github.com/GoFurry/fiberx/internal/planner"
	"github.com/GoFurry/fiberx/internal/renderer"
	"github.com/GoFurry/fiberx/internal/writer"
)

type Summary struct {
	Base            string
	PresetPacks     []string
	CapabilityPacks []string
	Preset          string
	Capabilities    []string
	ReplaceRules    []string
	InjectionRules  []string
	WrittenFiles    int
	WrittenPaths    []string
	Warnings        []string
	DryRun          bool
	TargetDir       string
}

func Build(plan planner.Plan, rendered renderer.Result, writeResult writer.Result) Summary {
	capabilities := make([]string, 0, len(plan.Capabilities))
	for _, capability := range plan.Capabilities {
		capabilities = append(capabilities, capability.Name)
	}

	presetPacks := make([]string, 0, len(plan.PresetPacks))
	for _, pack := range plan.PresetPacks {
		presetPacks = append(presetPacks, pack.Name)
	}

	capabilityPacks := make([]string, 0, len(plan.CapabilityPacks))
	for _, pack := range plan.CapabilityPacks {
		capabilityPacks = append(capabilityPacks, pack.Name)
	}

	return Summary{
		Base:            plan.Base.Name,
		PresetPacks:     presetPacks,
		CapabilityPacks: capabilityPacks,
		Preset:          plan.Preset.Name,
		Capabilities:    capabilities,
		ReplaceRules:    append([]string(nil), rendered.ReplaceRuleHits...),
		InjectionRules:  append([]string(nil), rendered.InjectionHits...),
		WrittenFiles:    writeResult.WrittenFiles,
		WrittenPaths:    append([]string(nil), writeResult.WrittenPaths...),
		Warnings:        append([]string(nil), rendered.Warnings...),
		DryRun:          writeResult.DryRun,
		TargetDir:       writeResult.TargetDir,
	}
}
