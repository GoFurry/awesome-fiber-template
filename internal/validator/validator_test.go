package validator

import (
	"strings"
	"testing"

	"github.com/GoFurry/fiberx/internal/manifest"
)

func TestValidateCatalogRejectsDuplicatePreset(t *testing.T) {
	catalog := manifest.Catalog{
		Presets: []manifest.PresetManifest{
			{Name: "light", Summary: "a", Description: "a", Base: "light"},
			{Name: "light", Summary: "b", Description: "b", Base: "light"},
		},
	}

	if err := ValidateCatalog(catalog); err == nil || !strings.Contains(err.Error(), `duplicate preset "light"`) {
		t.Fatalf("expected duplicate preset error, got %v", err)
	}
}

func TestValidateCatalogRejectsPresetReferencingUnknownCapability(t *testing.T) {
	catalog := manifest.Catalog{
		Presets: []manifest.PresetManifest{
			{Name: "light", Summary: "a", Description: "a", Base: "light", AllowedCapabilities: []string{"redis"}},
		},
	}

	if err := ValidateCatalog(catalog); err == nil || !strings.Contains(err.Error(), `unknown allowed capability "redis"`) {
		t.Fatalf("expected unknown capability error, got %v", err)
	}
}

func TestValidateCatalogRejectsCapabilityReferencingUnknownPreset(t *testing.T) {
	catalog := manifest.Catalog{
		Presets: []manifest.PresetManifest{
			{Name: "light", Summary: "a", Description: "a", Base: "light"},
		},
		Capabilities: []manifest.CapabilityManifest{
			{Name: "redis", Summary: "r", Description: "r", AllowedPresets: []string{"heavy"}},
		},
	}

	if err := ValidateCatalog(catalog); err == nil || !strings.Contains(err.Error(), `references unknown preset "heavy"`) {
		t.Fatalf("expected unknown preset error, got %v", err)
	}
}

func TestValidateCatalogRejectsRuleReferencingUnknownCapability(t *testing.T) {
	catalog := manifest.Catalog{
		Presets: []manifest.PresetManifest{
			{Name: "light", Summary: "a", Description: "a", Base: "light"},
		},
		Capabilities: []manifest.CapabilityManifest{
			{Name: "embedded-ui", Summary: "ui", Description: "ui", AllowedPresets: []string{"light"}},
		},
		ReplaceRules: []manifest.ReplaceRule{
			{
				Name:  "global",
				Scope: manifest.Scope{Capabilities: []string{"redis"}},
				Replacements: []manifest.Replacement{
					{Placeholder: "{{project_name}}", ValueFrom: "project_name"},
				},
			},
		},
	}

	if err := ValidateCatalog(catalog); err == nil || !strings.Contains(err.Error(), `references unknown capability "redis"`) {
		t.Fatalf("expected unknown capability in rule error, got %v", err)
	}
}

func TestValidateRequestRejectsInvalidPresetCapabilityCombination(t *testing.T) {
	catalog, err := manifest.LoadCatalog(filepathJoinGenerator())
	if err != nil {
		t.Fatalf("LoadCatalog() returned error: %v", err)
	}
	if err := ValidateCatalog(catalog); err != nil {
		t.Fatalf("ValidateCatalog() returned error: %v", err)
	}

	if err := ValidateRequest("demo", "github.com/example/demo", "extra-light", []string{"redis"}, catalog); err == nil || !strings.Contains(err.Error(), `not allowed for preset "extra-light"`) {
		t.Fatalf("expected invalid combination error, got %v", err)
	}
}

func filepathJoinGenerator() string {
	return "../../generator"
}
