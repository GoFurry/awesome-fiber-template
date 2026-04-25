package manifest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type AssetFile struct {
	SourcePath string
	OutputPath string
}

func CollectAssetFiles(root string) ([]AssetFile, error) {
	files := []AssetFile{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			if rel == "." {
				return nil
			}
			if rel == "snippets" || strings.HasPrefix(rel, "snippets"+string(filepath.Separator)) {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(rel, ".snippet") {
			return nil
		}

		outputPath := filepath.ToSlash(rel)
		if strings.HasSuffix(outputPath, ".tmpl") {
			outputPath = strings.TrimSuffix(outputPath, ".tmpl")
		}

		files = append(files, AssetFile{
			SourcePath: path,
			OutputPath: outputPath,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk asset files in %q: %w", root, err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].OutputPath < files[j].OutputPath
	})

	return files, nil
}

func SnippetExists(root string, snippetPath string) bool {
	path := filepath.Join(root, filepath.FromSlash(snippetPath))
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
