package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/GoFurry/awesome-go-template/fiber/v3/roof/env"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Orm = &orm{}
var once sync.Once
var registeredModels []any
var registeredModelTypes = map[string]struct{}{}
var registeredModelsMu sync.RWMutex

func initOrm() {
	Orm.loadDBConfig()
}

type orm struct {
	engine *gorm.DB
	driver string
}

func InitDatabaseOnStart(models ...any) error {
	cfg := env.GetServerConfig().DataBase
	if !cfg.Enabled {
		slog.Info("database service disabled by config, skip initialization")
		return nil
	}

	engine := Orm.DB()
	if engine == nil {
		return fmt.Errorf("database engine is nil")
	}

	if err := Orm.AutoMigrate(models); err != nil {
		return fmt.Errorf("auto migrate database failed: %w", err)
	}

	slog.Info("database service initialized", "driver", Orm.Driver())
	return nil
}

func (db *orm) loadDBConfig() {
	if db.engine != nil {
		return
	}

	cfg := env.GetServerConfig().DataBase
	dialector, driver, err := buildDialector(cfg)
	if err != nil {
		slog.Error("build database dialector failed", "error", err, "driver", cfg.DBType)
		os.Exit(1)
	}

	engine, err := gorm.Open(dialector)
	if err != nil {
		slog.Error("open database error", "error", err, "driver", driver)
		os.Exit(1)
	}

	sqlDB, err := engine.DB()
	if err != nil {
		slog.Error("get sql db instance failed", "error", err)
		os.Exit(1)
	}

	configurePool(sqlDB, driver)

	if err = sqlDB.Ping(); err != nil {
		slog.Error("ping database failed", "error", err, "driver", driver)
		os.Exit(1)
	}

	db.engine = engine
	db.driver = driver
	slog.Info("database connected", "driver", driver)
}

func (db *orm) DB() *gorm.DB {
	once.Do(initOrm)
	return db.engine
}

func (db *orm) Ready() bool {
	return db.engine != nil
}

func (db *orm) Driver() string {
	return db.driver
}

func (db *orm) AutoMigrate(models ...any) error {
	engine := db.DB()
	if engine == nil {
		return errors.New("database is not initialized")
	}

	targets := models
	if len(targets) == 0 {
		targets = RegisteredModels()
	}
	if len(targets) == 0 {
		slog.Info("no database models registered, skip auto migrate")
		return nil
	}

	return engine.AutoMigrate(targets...)
}

func (db *orm) Close() {
	if db.engine == nil {
		return
	}

	sqlDB, err := db.engine.DB()
	if err != nil {
		slog.Error("get sql db instance failed", "error", err)
		return
	}

	if err = sqlDB.Close(); err != nil {
		slog.Error("close database pool failed", "error", err)
		return
	}

	db.engine = nil
	db.driver = ""
	slog.Info("database pool closed")
}

func RegisterModels(models ...any) {
	registeredModelsMu.Lock()
	defer registeredModelsMu.Unlock()

	for _, model := range models {
		if model == nil {
			continue
		}

		typeName := modelTypeName(model)
		if _, exists := registeredModelTypes[typeName]; exists {
			continue
		}

		registeredModelTypes[typeName] = struct{}{}
		registeredModels = append(registeredModels, model)
	}
}

func RegisteredModels() []any {
	registeredModelsMu.RLock()
	defer registeredModelsMu.RUnlock()

	result := make([]any, len(registeredModels))
	copy(result, registeredModels)
	return result
}

func buildDialector(cfg env.DataBaseConfig) (gorm.Dialector, string, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.DBType))
	switch driver {
	case "", "postgres", "postgresql":
		return postgres.Open(buildPostgresDSN(cfg.Postgres)), "postgres", nil
	case "mysql":
		return mysql.Open(buildMySQLDSN(cfg.MySQL)), "mysql", nil
	case "sqlite":
		dsn, err := buildSQLiteDSN(cfg.SQLite)
		if err != nil {
			return nil, "", err
		}
		return sqlite.Open(dsn), "sqlite", nil
	default:
		return nil, "", fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}
}

func buildPostgresDSN(cfg env.SQLDataBaseConfig) string {
	if strings.TrimSpace(cfg.DSN) != "" {
		return cfg.DSN
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBName,
	)
}

func buildMySQLDSN(cfg env.SQLDataBaseConfig) string {
	if strings.TrimSpace(cfg.DSN) != "" {
		return cfg.DSN
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}

func buildSQLiteDSN(cfg env.SQLiteDataBaseConfig) (string, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		dsn = strings.TrimSpace(cfg.Path)
	}
	if dsn == "" {
		dsn = "./data/app.db"
	}

	if dsn == ":memory:" || strings.HasPrefix(dsn, "file:") {
		return dsn, nil
	}

	dir := filepath.Dir(dsn)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create sqlite directory failed: %w", err)
		}
	}

	return dsn, nil
}

func configurePool(sqlDB *sql.DB, driver string) {
	switch driver {
	case "sqlite":
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetConnMaxLifetime(0)
		sqlDB.SetConnMaxIdleTime(0)
	default:
		sqlDB.SetMaxIdleConns(100)
		sqlDB.SetMaxOpenConns(1000)
		sqlDB.SetConnMaxLifetime(60 * time.Second)
		sqlDB.SetConnMaxIdleTime(30 * time.Second)
	}
}

func modelTypeName(model any) string {
	t := reflect.TypeOf(model)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if pkgPath := t.PkgPath(); pkgPath != "" {
		return pkgPath + "." + t.Name()
	}
	return t.String()
}
