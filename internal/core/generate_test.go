package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAcceptsValidPhaseTwoRequest(t *testing.T) {
	req := Request{
		ProjectName: "demo",
		ModulePath:  "github.com/example/demo",
		Preset:      "medium",
		Capabilities: []string{
			"redis",
		},
		Options: map[string]string{
			"command":       "new",
			"manifest_root": "../../generator",
			"target_dir":    t.TempDir(),
		},
	}

	if err := Generate(req); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(req.Options["target_dir"], "main.go")); err != nil {
		t.Fatalf("expected generated main.go to exist: %v", err)
	}
}

func TestGenerateRejectsUnknownPreset(t *testing.T) {
	req := Request{
		ProjectName: "demo",
		ModulePath:  "github.com/example/demo",
		Preset:      "light",
		Capabilities: []string{
			"swagger",
		},
		Options: map[string]string{
			"manifest_root": "../../generator",
			"target_dir":    t.TempDir(),
		},
	}

	if err := Generate(req); err == nil {
		t.Fatal("Generate() expected an error for an invalid preset-capability combination")
	}
}
