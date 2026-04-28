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

func TestValidateCatalogRejectsPresetCapabilityAsymmetry(t *testing.T) {
	catalog := manifest.Catalog{
		Presets: []manifest.PresetManifest{
			{Name: "light", Summary: "a", Description: "a", Base: "light", AllowedCapabilities: []string{"swagger"}},
		},
		Capabilities: []manifest.CapabilityManifest{
			{Name: "swagger", Summary: "docs", Description: "docs", AllowedPresets: []string{"heavy"}},
		},
	}

	if err := ValidateCatalog(catalog); err == nil || !strings.Contains(err.Error(), `preset "light" allowed capability "swagger" must also reference preset "light"`) {
		t.Fatalf("expected preset/capability symmetry error, got %v", err)
	}
}

func TestValidateCatalogRejectsCapabilityPresetAsymmetry(t *testing.T) {
	catalog := manifest.Catalog{
		Presets: []manifest.PresetManifest{
			{Name: "heavy", Summary: "a", Description: "a", Base: "heavy", AllowedCapabilities: []string{}},
		},
		Capabilities: []manifest.CapabilityManifest{
			{Name: "redis", Summary: "r", Description: "r", AllowedPresets: []string{"heavy"}},
		},
	}

	if err := ValidateCatalog(catalog); err == nil || !strings.Contains(err.Error(), `capability "redis" allowed preset "heavy" must also reference capability "redis"`) {
		t.Fatalf("expected capability/preset symmetry error, got %v", err)
	}
}

func TestValidateCatalogRejectsRuleReferencingUnknownCapability(t *testing.T) {
	catalog := manifest.Catalog{
		Presets: []manifest.PresetManifest{
			{Name: "light", Summary: "a", Description: "a", Base: "light", AllowedCapabilities: []string{"embedded-ui"}},
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

	if err := ValidateRequest("demo", "github.com/example/demo", "extra-light", []string{"redis"}, map[string]string{}, catalog); err == nil || !strings.Contains(err.Error(), `not allowed for preset "extra-light"`) {
		t.Fatalf("expected invalid combination error, got %v", err)
	}
}

func TestValidateRequestRejectsExtraLightRuntimeOptions(t *testing.T) {
	catalog, err := manifest.LoadCatalog(filepathJoinGenerator())
	if err != nil {
		t.Fatalf("LoadCatalog() returned error: %v", err)
	}
	if err := ValidateCatalog(catalog); err != nil {
		t.Fatalf("ValidateCatalog() returned error: %v", err)
	}

	testCases := []struct {
		name    string
		options map[string]string
		want    string
	}{
		{name: "logger", options: map[string]string{"logger": "zap"}, want: `does not support logger option`},
		{name: "db", options: map[string]string{"db": "pgsql"}, want: `does not support db option`},
		{name: "data access", options: map[string]string{"data_access": "sqlx"}, want: `does not support data access option`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRequest("demo", "github.com/example/demo", "extra-light", nil, tc.options, catalog)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestValidateAssetsAcceptsImplementedSlice(t *testing.T) {
	catalog, err := manifest.LoadCatalog(filepathJoinGenerator())
	if err != nil {
		t.Fatalf("LoadCatalog() returned error: %v", err)
	}
	if err := ValidateCatalog(catalog); err != nil {
		t.Fatalf("ValidateCatalog() returned error: %v", err)
	}
	if err := ValidateAssets(filepathJoinGenerator(), catalog); err != nil {
		t.Fatalf("ValidateAssets() returned error: %v", err)
	}
}

func TestValidateCatalogAcceptsCurrentCapabilityBoundaries(t *testing.T) {
	catalog, err := manifest.LoadCatalog(filepathJoinGenerator())
	if err != nil {
		t.Fatalf("LoadCatalog() returned error: %v", err)
	}

	if err := ValidateCatalog(catalog); err != nil {
		t.Fatalf("ValidateCatalog() returned error: %v", err)
	}

	assertPresetCapabilityBoundary(t, catalog, "medium", []string{"swagger", "embedded-ui"}, []string{"redis", "swagger", "embedded-ui"})
	assertPresetCapabilityBoundary(t, catalog, "heavy", []string{"swagger", "embedded-ui"}, []string{"swagger", "embedded-ui", "redis"})
	assertPresetCapabilityBoundary(t, catalog, "light", []string{}, []string{"swagger", "embedded-ui"})
	assertPresetCapabilityBoundary(t, catalog, "extra-light", []string{}, []string{})

	assertCapabilityAllowedPresets(t, catalog, "swagger", []string{"heavy", "medium", "light"})
	assertCapabilityAllowedPresets(t, catalog, "embedded-ui", []string{"heavy", "medium", "light"})
	assertCapabilityAllowedPresets(t, catalog, "redis", []string{"heavy", "medium"})
}

func filepathJoinGenerator() string {
	return "../../generator"
}

func assertPresetCapabilityBoundary(t *testing.T, catalog manifest.Catalog, presetName string, wantDefaults []string, wantAllowed []string) {
	t.Helper()

	preset, ok := catalog.FindPreset(presetName)
	if !ok {
		t.Fatalf("expected preset %q to exist", presetName)
	}

	if strings.Join(preset.DefaultCapabilities, ",") != strings.Join(wantDefaults, ",") {
		t.Fatalf("preset %q defaults mismatch: got %v want %v", presetName, preset.DefaultCapabilities, wantDefaults)
	}
	if strings.Join(preset.AllowedCapabilities, ",") != strings.Join(wantAllowed, ",") {
		t.Fatalf("preset %q allowed mismatch: got %v want %v", presetName, preset.AllowedCapabilities, wantAllowed)
	}
}

func assertCapabilityAllowedPresets(t *testing.T, catalog manifest.Catalog, capabilityName string, want []string) {
	t.Helper()

	capability, ok := catalog.FindCapability(capabilityName)
	if !ok {
		t.Fatalf("expected capability %q to exist", capabilityName)
	}

	if strings.Join(capability.AllowedPresets, ",") != strings.Join(want, ",") {
		t.Fatalf("capability %q allowed presets mismatch: got %v want %v", capabilityName, capability.AllowedPresets, want)
	}
}
