package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type ruleEnvelope struct {
	Kind string `yaml:"kind"`
}

func LoadCatalog(root string) (Catalog, error) {
	presetFiles, err := yamlFiles(PresetsDir(root))
	if err != nil {
		return Catalog{}, err
	}

	capabilityFiles, err := yamlFiles(CapabilitiesDir(root))
	if err != nil {
		return Catalog{}, err
	}

	ruleFiles, err := yamlFiles(RulesDir(root))
	if err != nil {
		return Catalog{}, err
	}

	catalog := Catalog{
		Presets:        make([]PresetManifest, 0, len(presetFiles)),
		Capabilities:   make([]CapabilityManifest, 0, len(capabilityFiles)),
		ReplaceRules:   []ReplaceRule{},
		InjectionRules: []InjectionRule{},
	}

	for _, file := range presetFiles {
		var preset PresetManifest
		if err := decodeYAMLFile(file, &preset); err != nil {
			return Catalog{}, fmt.Errorf("load preset manifest %q: %w", file, err)
		}
		catalog.Presets = append(catalog.Presets, preset)
	}

	for _, file := range capabilityFiles {
		var capability CapabilityManifest
		if err := decodeYAMLFile(file, &capability); err != nil {
			return Catalog{}, fmt.Errorf("load capability manifest %q: %w", file, err)
		}
		catalog.Capabilities = append(catalog.Capabilities, capability)
	}

	for _, file := range ruleFiles {
		var envelope ruleEnvelope
		if err := decodeYAMLFile(file, &envelope); err != nil {
			return Catalog{}, fmt.Errorf("load rule manifest %q: %w", file, err)
		}

		switch envelope.Kind {
		case "replace":
			var rule ReplaceRule
			if err := decodeYAMLFile(file, &rule); err != nil {
				return Catalog{}, fmt.Errorf("load replace rule %q: %w", file, err)
			}
			catalog.ReplaceRules = append(catalog.ReplaceRules, rule)
		case "injection":
			var rule InjectionRule
			if err := decodeYAMLFile(file, &rule); err != nil {
				return Catalog{}, fmt.Errorf("load injection rule %q: %w", file, err)
			}
			catalog.InjectionRules = append(catalog.InjectionRules, rule)
		default:
			return Catalog{}, fmt.Errorf("load rule manifest %q: unsupported kind %q", file, envelope.Kind)
		}
	}

	return catalog, nil
}

func yamlFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %q: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".yaml" && filepath.Ext(name) != ".yml" {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}

	sort.Strings(files)
	return files, nil
}

func decodeYAMLFile(path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, dest); err != nil {
		return err
	}

	return nil
}
