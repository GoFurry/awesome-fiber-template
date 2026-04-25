package planner

import (
	"testing"

	"github.com/GoFurry/fiberx/internal/manifest"
	"github.com/GoFurry/fiberx/internal/validator"
)

func TestBuildPlanSelectsMediumRedisAssetsAndRules(t *testing.T) {
	root := "../../generator"
	catalog, err := manifest.LoadCatalog(root)
	if err != nil {
		t.Fatalf("LoadCatalog() returned error: %v", err)
	}
	if err := validator.ValidateCatalog(catalog); err != nil {
		t.Fatalf("ValidateCatalog() returned error: %v", err)
	}
	if err := validator.ValidateAssets(root, catalog); err != nil {
		t.Fatalf("ValidateAssets() returned error: %v", err)
	}

	plan := BuildPlan("demo", "github.com/example/demo", "medium", []string{"redis"}, map[string]string{"target_dir": t.TempDir()}, root, catalog)

	if plan.Base.Name != "service-base" {
		t.Fatalf("expected base service-base, got %q", plan.Base.Name)
	}
	if len(plan.PresetPacks) != 1 || plan.PresetPacks[0].Name != "preset-medium" {
		t.Fatalf("expected one preset pack preset-medium, got %#v", plan.PresetPacks)
	}
	if len(plan.CapabilityPacks) != 1 || plan.CapabilityPacks[0].Name != "redis" {
		t.Fatalf("expected one capability pack redis, got %#v", plan.CapabilityPacks)
	}
	if len(plan.ReplaceRules) != 1 {
		t.Fatalf("expected 1 replace rule, got %d", len(plan.ReplaceRules))
	}
	if len(plan.InjectionRules) != 1 {
		t.Fatalf("expected 1 injection rule, got %d", len(plan.InjectionRules))
	}
}
