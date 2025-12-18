package config

import (
	"path/filepath"
	"runtime"
)

// toRelativePath converts an absolute path to a relative path from the project root.
func toRelativePath(absPath string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return absPath
	}
	// Go up from internal/pkg/config to project root
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(file))))
	rel, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		return absPath
	}
	return rel
}
