package businesscard

import (
	"os"
	"path/filepath"
)

// ResolveRepoPath walks up from the working directory to find the repo root
// (directory containing go.mod) and joins the given path segments.
func ResolveRepoPath(parts ...string) string {
	dir, err := os.Getwd()
	if err != nil {
		return filepath.Join(parts...)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(append([]string{dir}, parts...)...)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join(parts...)
		}
		dir = parent
	}
}
