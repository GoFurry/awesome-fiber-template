package migrate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateCreateUpStatusVersionDown(t *testing.T) {
	tempDir := t.TempDir()
	migrationDir := filepath.Join(tempDir, "migrations")
	dbPath := filepath.Join(tempDir, "app.db")

	service, err := New(Config{
		Dialect:      "sqlite",
		DSN:          dbPath,
		Dir:          migrationDir,
		AllowMissing: true,
	})
	if err != nil {
		t.Fatalf("new migrate service failed: %v", err)
	}
	t.Cleanup(func() {
		if err := service.Close(); err != nil {
			t.Fatalf("close migrate service failed: %v", err)
		}
	})

	path, err := service.Create("create users table", MigrationKindSQL)
	if err != nil {
		t.Fatalf("create migration failed: %v", err)
	}
	if !strings.HasSuffix(path, ".sql") {
		t.Fatalf("expected sql migration file, got %s", path)
	}

	sqlBody := `-- +goose Up
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
`
	if err := os.WriteFile(path, []byte(sqlBody), 0o644); err != nil {
		t.Fatalf("write migration body failed: %v", err)
	}

	if err := service.Up(context.Background()); err != nil {
		t.Fatalf("up migration failed: %v", err)
	}

	version, err := service.Version(context.Background())
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if version == 0 {
		t.Fatalf("expected non-zero version after up")
	}

	statuses, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected one migration status, got %d", len(statuses))
	}
	if statuses[0].State == "" {
		t.Fatalf("expected migration state to be populated")
	}

	if err := service.Down(context.Background()); err != nil {
		t.Fatalf("down migration failed: %v", err)
	}

	version, err = service.Version(context.Background())
	if err != nil {
		t.Fatalf("version after down failed: %v", err)
	}
	if version != 0 {
		t.Fatalf("expected version to return to zero, got %d", version)
	}
}

func TestMigrateTableOverride(t *testing.T) {
	tempDir := t.TempDir()
	migrationDir := filepath.Join(tempDir, "migrations")
	dbPath := filepath.Join(tempDir, "app.db")

	service, err := New(Config{
		Dialect:      "sqlite",
		DSN:          dbPath,
		Dir:          migrationDir,
		Table:        "custom_schema_migrations",
		AllowMissing: true,
	})
	if err != nil {
		t.Fatalf("new migrate service failed: %v", err)
	}
	t.Cleanup(func() {
		if err := service.Close(); err != nil {
			t.Fatalf("close migrate service failed: %v", err)
		}
	})

	if service.cfg.Table != "custom_schema_migrations" {
		t.Fatalf("expected custom table name to be preserved")
	}
}

func TestMigrateValidationErrors(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatalf("expected validation error")
	}

	tempDir := t.TempDir()
	if _, err := New(Config{
		Dialect: "unknown",
		DSN:     "app.db",
		Dir:     filepath.Join(tempDir, "migrations"),
	}); err == nil {
		t.Fatalf("expected unsupported dialect error")
	}

	if _, err := New(Config{
		Dialect: "sqlite",
		DSN:     filepath.Join(tempDir, "app.db"),
		Dir:     filepath.Join(tempDir, "missing"),
	}); err == nil {
		t.Fatalf("expected missing dir error when allow missing is false")
	}
}
