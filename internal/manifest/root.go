package manifest

import (
	"os"
	"path/filepath"
)

func ResolveRoot(explicit string) string {
	if explicit != "" {
		return explicit
	}

	if envRoot := os.Getenv("FIBERX_MANIFEST_ROOT"); envRoot != "" {
		return envRoot
	}

	for _, candidate := range candidateRoots() {
		if generatorRootExists(candidate) {
			return candidate
		}
	}

	return DefaultRoot()
}

func candidateRoots() []string {
	candidates := []string{DefaultRoot()}

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, DefaultRoot()))
	}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(dir, DefaultRoot()))
		candidates = append(candidates, filepath.Join(filepath.Dir(dir), DefaultRoot()))
	}

	return candidates
}

func generatorRootExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
