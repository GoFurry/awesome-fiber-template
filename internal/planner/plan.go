package planner

import "github.com/GoFurry/fiberx/internal/manifest"

type AssetSelection struct {
	Kind string
	Name string
	Dir  string
}

type Plan struct {
	ProjectName     string
	ModulePath      string
	TargetDir       string
	Preset          manifest.PresetManifest
	Capabilities    []manifest.CapabilityManifest
	Base            AssetSelection
	PresetPacks     []AssetSelection
	CapabilityPacks []AssetSelection
	Assets          []AssetSelection
	ReplaceRules    []manifest.ReplaceRule
	InjectionRules  []manifest.InjectionRule
	Options         map[string]string
}

func BuildPlan(projectName string, modulePath string, presetName string, capabilityNames []string, options map[string]string, catalogRoot string, catalog manifest.Catalog) Plan {
	preset, _ := catalog.FindPreset(presetName)

	selectedCapabilityNames := mergeCapabilityNames(catalog.AppliedDefaultCapabilities(preset), capabilityNames)
	capabilities := make([]manifest.CapabilityManifest, 0, len(selectedCapabilityNames))
	for _, name := range selectedCapabilityNames {
		capability, ok := catalog.FindCapability(name)
		if !ok {
			continue
		}
		capabilities = append(capabilities, capability)
	}

	base := AssetSelection{
		Kind: "base",
		Name: preset.Base,
		Dir:  manifest.BaseAssetDir(catalogRoot, preset.Base),
	}

	presetPacks := make([]AssetSelection, 0, len(preset.Packs))
	for _, pack := range preset.Packs {
		presetPacks = append(presetPacks, AssetSelection{
			Kind: "preset-pack",
			Name: pack,
			Dir:  manifest.PackAssetDir(catalogRoot, pack),
		})
	}

	capabilityPacks := make([]AssetSelection, 0)
	for _, capability := range capabilities {
		for _, pack := range capability.Packs {
			capabilityPacks = append(capabilityPacks, AssetSelection{
				Kind: "capability-pack",
				Name: pack,
				Dir:  manifest.CapabilityAssetDir(catalogRoot, pack),
			})
		}
	}

	assets := make([]AssetSelection, 0, 1+len(presetPacks)+len(capabilityPacks))
	if base.Name != "" {
		assets = append(assets, base)
	}
	assets = append(assets, presetPacks...)
	assets = append(assets, capabilityPacks...)

	return Plan{
		ProjectName:     projectName,
		ModulePath:      modulePath,
		TargetDir:       options["target_dir"],
		Preset:          preset,
		Capabilities:    capabilities,
		Base:            base,
		PresetPacks:     presetPacks,
		CapabilityPacks: capabilityPacks,
		Assets:          assets,
		ReplaceRules:    selectReplaceRules(catalog.ReplaceRules, preset.Name, selectedCapabilityNames),
		InjectionRules:  selectInjectionRules(catalog.InjectionRules, preset.Name, selectedCapabilityNames),
		Options:         cloneOptions(options),
	}
}

func cloneOptions(options map[string]string) map[string]string {
	if len(options) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(options))
	for key, value := range options {
		cloned[key] = value
	}

	return cloned
}

func mergeCapabilityNames(defaults []string, requested []string) []string {
	if len(defaults) == 0 && len(requested) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(defaults)+len(requested))
	merged := make([]string, 0, len(defaults)+len(requested))
	for _, name := range append(append([]string{}, defaults...), requested...) {
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		merged = append(merged, name)
	}

	return merged
}

func selectReplaceRules(rules []manifest.ReplaceRule, presetName string, capabilityNames []string) []manifest.ReplaceRule {
	selected := make([]manifest.ReplaceRule, 0, len(rules))
	for _, rule := range rules {
		if !matchesScope(rule.Scope, presetName, capabilityNames) {
			continue
		}
		selected = append(selected, rule)
	}
	return selected
}

func selectInjectionRules(rules []manifest.InjectionRule, presetName string, capabilityNames []string) []manifest.InjectionRule {
	selected := make([]manifest.InjectionRule, 0, len(rules))
	for _, rule := range rules {
		if !matchesScope(rule.Scope, presetName, capabilityNames) {
			continue
		}
		selected = append(selected, rule)
	}
	return selected
}

func matchesScope(scope manifest.Scope, presetName string, capabilityNames []string) bool {
	if len(scope.Presets) > 0 && !contains(scope.Presets, presetName) {
		return false
	}

	if len(scope.Capabilities) > 0 {
		matched := false
		for _, capability := range scope.Capabilities {
			if contains(capabilityNames, capability) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
