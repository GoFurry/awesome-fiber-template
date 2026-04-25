package renderer

import (
	"strings"
	"testing"

	"github.com/GoFurry/fiberx/internal/manifest"
	"github.com/GoFurry/fiberx/internal/planner"
	"github.com/GoFurry/fiberx/internal/validator"
)

func TestRenderAppliesReplacementsAndInjection(t *testing.T) {
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

	plan := planner.BuildPlan("demo", "github.com/example/demo", "medium", []string{"redis"}, map[string]string{"target_dir": t.TempDir()}, root, catalog)
	result, err := Render(plan)
	if err != nil {
		t.Fatalf("Render() returned error: %v", err)
	}

	bootstrap := findRenderedFile(t, result, "internal/bootstrap/bootstrap.go")
	if !strings.Contains(bootstrap, `services = append(services, "cache:redis")`) {
		t.Fatalf("expected redis injection in bootstrap.go, got:\n%s", bootstrap)
	}

	goMod := findRenderedFile(t, result, "go.mod")
	if !strings.Contains(goMod, "module github.com/example/demo") {
		t.Fatalf("expected rendered go.mod module path, got:\n%s", goMod)
	}
}

func findRenderedFile(t *testing.T, result Result, path string) string {
	t.Helper()

	for _, file := range result.Files {
		if file.Path == path {
			return string(file.Content)
		}
	}

	t.Fatalf("rendered file %q not found", path)
	return ""
}
