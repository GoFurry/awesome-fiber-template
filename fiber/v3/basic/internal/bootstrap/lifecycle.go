package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/basic/config"
	cache "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/cache"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/db"
	log "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/logging"
	scheduler "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/scheduler"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/schedule"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/transport/http/middleware"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/pkg/common"
	corazalite "github.com/GoFurry/coraza-fiber-lite"
)

var (
	lifecycleMu sync.Mutex
	started     bool
	currentApp  *Application
)

func Start() (*Application, error) {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if started {
		return currentApp, nil
	}

	cfg := env.GetServerConfig()

	if err := initLogger(cfg); err != nil {
		return nil, err
	}

	cleanupOnError := func(cause error) (*Application, error) {
		return nil, errors.Join(cause, shutdownComponents(cfg))
	}

	if cfg.Prometheus.Enabled {
		middleware.InitPrometheus(middleware.FiberPromConf{
			Namespace:         cfg.Prometheus.Namespace,
			SkipPaths:         []string{cfg.Prometheus.Path},
			IgnoreStatusCodes: []int{},
		})
	}

	if cfg.Waf.Enabled {
		if err := initWAF(cfg); err != nil {
			return cleanupOnError(err)
		}
	}

	if cfg.DataBase.Enabled {
		if err := db.InitDatabaseOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("database init failed: %w", err))
		}
	}

	if cfg.Redis.Enabled {
		if err := cache.InitRedisOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("redis init failed: %w", err))
		}
	}

	if cfg.Schedule.Enabled {
		if err := scheduler.InitTimeWheelOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("scheduler init failed: %w", err))
		}
		if err := schedule.InitScheduleOnStart(); err != nil {
			return cleanupOnError(fmt.Errorf("schedule init failed: %w", err))
		}
	}

	app, err := buildApplication()
	if err != nil {
		return cleanupOnError(fmt.Errorf("build application failed: %w", err))
	}

	currentApp = app
	started = true
	slog.Info("application bootstrap completed")
	return currentApp, nil
}

func Shutdown() error {
	lifecycleMu.Lock()
	defer lifecycleMu.Unlock()

	if !started {
		return nil
	}

	cfg := env.GetServerConfig()
	err := shutdownComponents(cfg)
	started = false
	currentApp = nil
	return err
}

func initLogger(cfg *env.ServerConfigHolder) error {
	logCfg := &log.Config{
		ShowLine:   true,
		TimeFormat: common.TIME_FORMAT_DATE,
	}

	if cfg.Server.Mode == "debug" {
		logCfg.Level = "debug"
		logCfg.Mode = "dev"
		logCfg.EncodeJson = false
	} else {
		logCfg.Level = cfg.Log.LogLevel
		logCfg.Mode = cfg.Log.LogMode
		logCfg.FilePath = cfg.Log.LogPath
		logCfg.MaxSize = cfg.Log.LogMaxSize
		logCfg.MaxBackups = cfg.Log.LogMaxBackups
		logCfg.MaxAge = cfg.Log.LogMaxAge
		logCfg.Compress = true
		logCfg.EncodeJson = true
		logCfg.TimeFormat = common.TIME_FORMAT_LOG
	}

	if err := log.InitLogger(logCfg); err != nil {
		return fmt.Errorf("logger init failed: %w", err)
	}
	return nil
}

func initWAF(cfg *env.ServerConfigHolder) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("waf init panic: %v", recovered)
		}
	}()

	corazalite.InitMetrics()
	corazalite.InitGlobalWAFWithCfg(corazalite.CorazaCfg{
		DirectivesFile:     cfg.Waf.ConfPath,
		RequestBodyAccess:  true,
		ResponseBodyAccess: false,
	})
	corazalite.InitWAFBlockMessage("Request blocked by CorazaLite WAF")
	return nil
}

func shutdownComponents(cfg *env.ServerConfigHolder) error {
	var shutdownErr error

	if cfg.Schedule.Enabled {
		scheduler.Stop()
	}

	if cfg.Redis.Enabled {
		if err := cache.Close(); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("redis shutdown failed: %w", err))
		}
	}

	if cfg.DataBase.Enabled {
		db.Orm.Close()
	}

	if err := log.Sync(); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("logger sync failed: %w", err))
	}

	return shutdownErr
}
