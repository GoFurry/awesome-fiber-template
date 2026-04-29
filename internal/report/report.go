package report

import (
	"github.com/GoFurry/fiberx/internal/planner"
	"github.com/GoFurry/fiberx/internal/renderer"
	"github.com/GoFurry/fiberx/internal/writer"
)

type Summary struct {
	Base                      string
	FiberVersion              string
	CLIStyle                  string
	Logger                    string
	Database                  string
	DataAccess                string
	GeneratorVersion          string
	GeneratorCommit           string
	TemplateSetFingerprint    string
	RenderedOutputFingerprint string
	MetadataPath              string
	PresetPacks               []string
	CapabilityPacks           []string
	RuntimeOverlays           []string
	Preset                    string
	Capabilities              []string
	ReplaceRules              []string
	InjectionRules            []string
	WrittenFiles              int
	WrittenPaths              []string
	Warnings                  []string
	DryRun                    bool
	TargetDir                 string
}

func Build(plan planner.Plan, rendered renderer.Result, writeResult writer.Result, generatorVersion, generatorCommit, templateSetFingerprint, renderedOutputFingerprint, metadataPath string) Summary {
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

	runtimeOverlays := make([]string, 0, len(plan.RuntimeOverlays))
	for _, pack := range plan.RuntimeOverlays {
		runtimeOverlays = append(runtimeOverlays, pack.Name)
	}

	return Summary{
		Base:                      plan.Base.Name,
		FiberVersion:              plan.FiberVersion,
		CLIStyle:                  plan.CLIStyle,
		Logger:                    plan.Logger,
		Database:                  plan.Database,
		DataAccess:                plan.DataAccess,
		GeneratorVersion:          generatorVersion,
		GeneratorCommit:           generatorCommit,
		TemplateSetFingerprint:    templateSetFingerprint,
		RenderedOutputFingerprint: renderedOutputFingerprint,
		MetadataPath:              metadataPath,
		PresetPacks:               presetPacks,
		CapabilityPacks:           capabilityPacks,
		RuntimeOverlays:           runtimeOverlays,
		Preset:                    plan.Preset.Name,
		Capabilities:              capabilities,
		ReplaceRules:              append([]string(nil), rendered.ReplaceRuleHits...),
		InjectionRules:            append([]string(nil), rendered.InjectionHits...),
		WrittenFiles:              writeResult.WrittenFiles,
		WrittenPaths:              append([]string(nil), writeResult.WrittenPaths...),
		Warnings:                  append([]string(nil), rendered.Warnings...),
		DryRun:                    writeResult.DryRun,
		TargetDir:                 writeResult.TargetDir,
	}
}
