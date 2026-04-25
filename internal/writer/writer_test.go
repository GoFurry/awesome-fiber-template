package writer

import (
	"testing"

	"github.com/GoFurry/fiberx/internal/renderer"
)

func TestWriteRefusesOverwrite(t *testing.T) {
	targetDir := t.TempDir()
	rendered := renderer.Result{
		Files: []renderer.File{
			{Path: "main.go", Content: []byte("package main\n")},
		},
	}

	if _, err := New(targetDir).Write(rendered); err != nil {
		t.Fatalf("first Write() returned error: %v", err)
	}

	if _, err := New(targetDir).Write(rendered); err == nil {
		t.Fatal("expected second Write() to fail on overwrite")
	}
}
