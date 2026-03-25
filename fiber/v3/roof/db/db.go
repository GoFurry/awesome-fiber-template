package db

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/GoFurry/awesome-go-template/fiber/v3/roof/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Orm = &orm{}
var once sync.Once

func initOrm() {
	Orm.loadDBConfig()
}

type orm struct {
	engine *gorm.DB
}

func (db *orm) loadDBConfig() {
	if db.engine != nil {
		return
	}

	pgsql := env.GetServerConfig().DataBase
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgsql.DBHost,
		pgsql.DBPort,
		pgsql.DBUsername,
		pgsql.DBPassword,
		pgsql.DBName,
	)

	engine, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		slog.Error("open database error", "error", err)
		os.Exit(1)
	}

	sqlDB, err := engine.DB()
	if err != nil {
		slog.Error("get sql db instance failed", "error", err)
		os.Exit(1)
	}

	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(1000)
	sqlDB.SetConnMaxLifetime(60 * time.Second)
	sqlDB.SetConnMaxIdleTime(30 * time.Second)

	if err = sqlDB.Ping(); err != nil {
		slog.Error("ping database failed", "error", err)
		os.Exit(1)
	}

	db.engine = engine
}

func (db *orm) DB() *gorm.DB {
	once.Do(initOrm)
	return db.engine
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
	slog.Info("database pool closed")
}
