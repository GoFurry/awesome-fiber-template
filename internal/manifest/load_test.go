package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCatalogLoadsDiskDeclarations(t *testing.T) {
	catalog, err := LoadCatalog(filepath.Join("..", "..", "generator"))
	if err != nil {
		t.Fatalf("LoadCatalog() returned error: %v", err)
	}

	if len(catalog.Presets) != 4 {
		t.Fatalf("expected 4 presets, got %d", len(catalog.Presets))
	}

	if len(catalog.Capabilities) != 3 {
		t.Fatalf("expected 3 capabilities, got %d", len(catalog.Capabilities))
	}

	if len(catalog.ReplaceRules) == 0 {
		t.Fatal("expected at least one replace rule")
	}

	if len(catalog.InjectionRules) == 0 {
		t.Fatal("expected at least one injection rule")
	}
}

func TestLoadCatalogRejectsUnknownRuleKind(t *testing.T) {
	root := writeCatalogFixture(t, map[string]string{
		"presets/light.yaml": `name: light
summary: light
description: light
base: light
default_capabilities: []
allowed_capabilities: []
`,
		"capabilities/redis.yaml": `name: redis
summary: redis
description: redis
allowed_presets: [light]
depends_on: []
conflicts_with: []
`,
		"rules/bad.yaml": `kind: mystery
name: bad
scope: {}
`,
	})

	if _, err := LoadCatalog(root); err == nil {
		t.Fatal("expected LoadCatalog() to fail for an unknown rule kind")
	}
}

func writeCatalogFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for relative, content := range files {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) failed: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) failed: %v", path, err)
		}
	}

	return root
}
