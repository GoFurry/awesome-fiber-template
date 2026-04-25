package manifest

import "path/filepath"

type Catalog struct {
	Presets        []PresetManifest
	Capabilities   []CapabilityManifest
	ReplaceRules   []ReplaceRule
	InjectionRules []InjectionRule
}

type PresetManifest struct {
	Name                string   `yaml:"name"`
	Summary             string   `yaml:"summary"`
	Description         string   `yaml:"description"`
	Base                string   `yaml:"base"`
	DefaultCapabilities []string `yaml:"default_capabilities"`
	AllowedCapabilities []string `yaml:"allowed_capabilities"`
}

type CapabilityManifest struct {
	Name           string   `yaml:"name"`
	Summary        string   `yaml:"summary"`
	Description    string   `yaml:"description"`
	AllowedPresets []string `yaml:"allowed_presets"`
	DependsOn      []string `yaml:"depends_on"`
	ConflictsWith  []string `yaml:"conflicts_with"`
}

type Scope struct {
	Presets      []string `yaml:"presets"`
	Capabilities []string `yaml:"capabilities"`
}

type Replacement struct {
	Placeholder string `yaml:"placeholder"`
	ValueFrom   string `yaml:"value_from"`
}

type ReplaceRule struct {
	Name         string        `yaml:"name"`
	Scope        Scope         `yaml:"scope"`
	Replacements []Replacement `yaml:"replacements"`
}

type InjectionRule struct {
	Name    string `yaml:"name"`
	Scope   Scope  `yaml:"scope"`
	Target  string `yaml:"target"`
	Anchor  string `yaml:"anchor"`
	Snippet string `yaml:"snippet"`
	Order   int    `yaml:"order"`
}

func DefaultRoot() string {
	return "generator"
}

func PresetsDir(root string) string {
	return filepath.Join(root, "presets")
}

func CapabilitiesDir(root string) string {
	return filepath.Join(root, "capabilities")
}

func RulesDir(root string) string {
	return filepath.Join(root, "rules")
}

func (c Catalog) FindPreset(name string) (PresetManifest, bool) {
	for _, preset := range c.Presets {
		if preset.Name == name {
			return preset, true
		}
	}

	return PresetManifest{}, false
}

func (c Catalog) FindCapability(name string) (CapabilityManifest, bool) {
	for _, capability := range c.Capabilities {
		if capability.Name == name {
			return capability, true
		}
	}

	return CapabilityManifest{}, false
}

func (c Catalog) HasCapability(name string) bool {
	_, ok := c.FindCapability(name)
	return ok
}

func (c Catalog) AppliedDefaultCapabilities(preset PresetManifest) []string {
	if len(preset.DefaultCapabilities) == 0 {
		return []string{}
	}

	capabilities := make([]string, 0, len(preset.DefaultCapabilities))
	for _, name := range preset.DefaultCapabilities {
		capability, ok := c.FindCapability(name)
		if !ok {
			continue
		}
		capabilities = append(capabilities, capability.Name)
	}

	return capabilities
}
