package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

const defaultTable = "schema_migrations"

type MigrationKind string

const MigrationKindSQL MigrationKind = "sql"

type Config struct {
	Dialect      string
	DSN          string
	Dir          string
	Table        string
	AllowMissing bool
	Verbose      bool
}

type MigrationStatus struct {
	Version   int64
	Name      string
	State     string
	AppliedAt time.Time
}

type Service struct {
	cfg      Config
	db       *sql.DB
	provider *goose.Provider
}

func New(cfg Config) (*Service, error) {
	normalized, driverName, _, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	if err := ensureDir(normalized.Dir, normalized.AllowMissing); err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, normalized.DSN)
	if err != nil {
		return nil, fmt.Errorf("open migration database failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping migration database failed: %w", err)
	}

	return &Service{
		cfg: normalized,
		db:  db,
	}, nil
}

func (s *Service) Close() error {
	if s == nil {
		return nil
	}

	if s.db != nil {
		if s.provider != nil {
			closeErr := s.provider.Close()
			s.provider = nil
			s.db = nil
			return closeErr
		}

		closeErr := s.db.Close()
		s.db = nil
		return closeErr
	}

	s.provider = nil
	return nil
}

func (s *Service) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	provider, err := s.getProvider()
	if errors.Is(err, goose.ErrNoMigrations) {
		return []MigrationStatus{}, nil
	}
	if err != nil {
		return nil, err
	}

	statuses, err := provider.Status(ctxOrBackground(ctx))
	if err != nil {
		return nil, err
	}

	results := make([]MigrationStatus, 0, len(statuses))
	for _, status := range statuses {
		if status == nil || status.Source == nil {
			continue
		}
		results = append(results, MigrationStatus{
			Version:   status.Source.Version,
			Name:      filepath.Base(status.Source.Path),
			State:     string(status.State),
			AppliedAt: status.AppliedAt,
		})
	}
	return results, nil
}

func (s *Service) Up(ctx context.Context) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	provider, err := s.getProvider()
	if errors.Is(err, goose.ErrNoMigrations) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = provider.Up(ctxOrBackground(ctx))
	return err
}

func (s *Service) Down(ctx context.Context) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	provider, err := s.getProvider()
	if errors.Is(err, goose.ErrNoMigrations) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = provider.Down(ctxOrBackground(ctx))
	return err
}

func (s *Service) Create(name string, kind MigrationKind) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	if kind != MigrationKindSQL {
		return "", fmt.Errorf("migration kind %q is unsupported", kind)
	}

	if err := os.MkdirAll(s.cfg.Dir, 0o755); err != nil {
		return "", fmt.Errorf("create migration dir failed: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.sql", time.Now().UTC().Format("20060102150405"), sanitizeName(name))
	path := filepath.Join(s.cfg.Dir, filename)
	body := []byte(defaultSQLTemplate)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", fmt.Errorf("write migration file failed: %w", err)
	}
	s.provider = nil
	return path, nil
}

func (s *Service) Version(ctx context.Context) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}

	provider, err := s.getProvider()
	if errors.Is(err, goose.ErrNoMigrations) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return provider.GetDBVersion(ctxOrBackground(ctx))
}

func (s *Service) ensureReady() error {
	if s == nil || s.db == nil {
		return errors.New("migration service is not initialized")
	}
	return nil
}

func (s *Service) getProvider() (*goose.Provider, error) {
	if s.provider != nil {
		return s.provider, nil
	}

	_, dialect, err := resolveDialect(s.cfg.Dialect)
	if err != nil {
		return nil, err
	}

	provider, err := goose.NewProvider(
		dialect,
		s.db,
		os.DirFS(s.cfg.Dir),
		goose.WithTableName(s.cfg.Table),
		goose.WithVerbose(s.cfg.Verbose),
	)
	if err != nil {
		return nil, fmt.Errorf("create goose provider failed: %w", err)
	}
	s.provider = provider
	return s.provider, nil
}

func normalizeConfig(cfg Config) (Config, string, goose.Dialect, error) {
	normalized := cfg
	normalized.DSN = strings.TrimSpace(normalized.DSN)
	normalized.Dir = strings.TrimSpace(normalized.Dir)
	normalized.Table = strings.TrimSpace(normalized.Table)

	if normalized.DSN == "" {
		return Config{}, "", "", errors.New("migration dsn is required")
	}
	if normalized.Dir == "" {
		return Config{}, "", "", errors.New("migration dir is required")
	}
	if normalized.Table == "" {
		normalized.Table = defaultTable
	}

	driverName, dialect, err := resolveDialect(strings.TrimSpace(normalized.Dialect))
	if err != nil {
		return Config{}, "", "", err
	}
	normalized.Dialect = string(dialect)
	return normalized, driverName, dialect, nil
}

func resolveDialect(input string) (string, goose.Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "sqlite", "sqlite3":
		return "sqlite", goose.Dialect("sqlite3"), nil
	case "postgres", "postgresql", "pgx":
		return "pgx", goose.Dialect("postgres"), nil
	case "mysql":
		return "mysql", goose.Dialect("mysql"), nil
	default:
		return "", "", fmt.Errorf("migration dialect %q is unsupported", input)
	}
}

func ensureDir(dir string, allowMissing bool) error {
	info, err := os.Stat(dir)
	switch {
	case err == nil && info.IsDir():
		return nil
	case err == nil && !info.IsDir():
		return fmt.Errorf("migration dir is not a directory: %s", dir)
	case errors.Is(err, os.ErrNotExist) && allowMissing:
		return os.MkdirAll(dir, 0o755)
	case errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("migration dir does not exist: %s", dir)
	default:
		return fmt.Errorf("stat migration dir failed: %w", err)
	}
}

var nonSlugChar = regexp.MustCompile(`[^a-z0-9_]+`)

func sanitizeName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = nonSlugChar.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		return "new_migration"
	}
	return normalized
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

const defaultSQLTemplate = `-- +goose Up
-- TODO: write your SQL here

-- +goose Down
-- TODO: write your rollback SQL here
`
