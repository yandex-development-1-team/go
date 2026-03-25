package database

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveMigrationsDir_DefaultFromGoMod(t *testing.T) {
	t.Parallel()

	dir, err := ResolveMigrationsDir("")
	if err != nil {
		t.Fatalf("ResolveMigrationsDir(\"\"): %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(dir), "/migrations") {
		t.Fatalf("expected path ending with /migrations, got %q", dir)
	}
}

func TestResolveMigrationsDir_ExplicitSameAsDefault(t *testing.T) {
	t.Parallel()

	defaultDir, err := ResolveMigrationsDir("")
	if err != nil {
		t.Fatalf("default: %v", err)
	}
	again, err := ResolveMigrationsDir(defaultDir)
	if err != nil {
		t.Fatalf("explicit default path: %v", err)
	}
	if filepath.Clean(again) != filepath.Clean(defaultDir) {
		t.Fatalf("got %q want %q", again, defaultDir)
	}
}

func TestResolveMigrationsDir_ExplicitMissing(t *testing.T) {
	t.Parallel()

	_, err := ResolveMigrationsDir("/no/such/migrations/dir/12345")
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}
