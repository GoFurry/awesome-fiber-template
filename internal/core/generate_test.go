package core

import "testing"

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
		},
	}

	if err := Generate(req); err != nil {
		t.Fatalf("Generate() returned error: %v", err)
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
		},
	}

	if err := Generate(req); err == nil {
		t.Fatal("Generate() expected an error for an invalid preset-capability combination")
	}
}
