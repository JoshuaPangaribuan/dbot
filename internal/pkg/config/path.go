package config

import (
	"os"
	"path/filepath"
)

// toRelativePath converts an absolute path to a relative path from cwd.
// This is used for error messages to be more user-friendly.
func toRelativePath(absPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return absPath
	}
	rel, err := filepath.Rel(cwd, absPath)
	if err != nil {
		return absPath
	}
	return rel
}
