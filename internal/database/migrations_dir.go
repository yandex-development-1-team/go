package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultMigrationsDirName = "migrations"

func ResolveMigrationsDir(explicit string) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", fmt.Errorf("migrations path %q: %w", explicit, err)
		}
		st, err := os.Stat(abs)
		if err != nil {
			return "", fmt.Errorf("migrations dir %q: %w", abs, err)
		}
		if !st.IsDir() {
			return "", fmt.Errorf("migrations path %q is not a directory", abs)
		}
		return abs, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	if root, ok := findGoModDir(wd); ok {
		dir := filepath.Join(root, DefaultMigrationsDirName)
		abs, err := filepath.Abs(dir)
		if err != nil {
			return "", err
		}
		st, err := os.Stat(abs)
		if err != nil {
			return "", fmt.Errorf("migrations dir %q (next to go.mod): %w", abs, err)
		}
		if !st.IsDir() {
			return "", fmt.Errorf("migrations path %q is not a directory", abs)
		}
		return abs, nil
	}

	abs, err := filepath.Abs(DefaultMigrationsDirName)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("migrations dir %q (relative to working directory): %w; set MIGRATIONS_DIR or run with module root as cwd", abs, err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("migrations path %q is not a directory", abs)
	}
	return abs, nil
}

func findGoModDir(start string) (string, bool) {
	dir := start
	for {
		st, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil && !st.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
