package postprocess

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func FinalizeGeneratedModule(targetDir string) error {
	if targetDir == "" {
		return nil
	}

	goModPath := filepath.Join(targetDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat generated go.mod: %w", err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("finalize generated module with go mod tidy: %w\n%s", err, string(output))
	}

	return nil
}
