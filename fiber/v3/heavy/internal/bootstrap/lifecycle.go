package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/config"
	cache "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/infra/cache"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/infra/db"
	log "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/infra/logging"
	scheduler "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/infra/scheduler"
	modules "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/pkg/common"
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

	var app *Application
	cleanupOnError := func(cause error) (*Application, error) {
		return nil, errors.Join(cause, shutdownComponents(cfg, app))
	}

	var err error
	app, err = buildApplication()
	if err != nil {
		return cleanupOnError(fmt.Errorf("build application failed: %w", err))
	}

	if cfg.Waf.Enabled {
		if err := initWAF(cfg); err != nil {
			return cleanupOnError(err)
		}
	}

	if cfg.DataBase.Enabled {
		if err := db.InitDatabaseOnStart(app.DatabaseModels...); err != nil {
			return cleanupOnError(fmt.Errorf("database init failed: %w", err))
		}
		if err := db.ApplyMigrations(app.Migrations...); err != nil {
			return cleanupOnError(fmt.Errorf("database migrations failed: %w", err))
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
		if err := registerScheduledJobs(app.ScheduledJobs); err != nil {
			return cleanupOnError(fmt.Errorf("schedule registration failed: %w", err))
		}
	}

	if err := startBackgroundServices(app.BackgroundServices); err != nil {
		return cleanupOnError(fmt.Errorf("background service start failed: %w", err))
	}

	if err := runHooks(app.StartupHooks); err != nil {
		return cleanupOnError(fmt.Errorf("startup hook failed: %w", err))
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
	err := shutdownComponents(cfg, currentApp)
	started = false
	currentApp = nil
	return err
}

func RunMigrations() error {
	app, err := buildApplication()
	if err != nil {
		return fmt.Errorf("build application failed: %w", err)
	}

	if err := db.InitDatabaseOnStart(app.DatabaseModels...); err != nil {
		return fmt.Errorf("database init failed: %w", err)
	}
	if err := db.ApplyMigrations(app.Migrations...); err != nil {
		return fmt.Errorf("database migrations failed: %w", err)
	}
	return nil
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
	corazalite.InitWAFBlockMessage("Request blocked by CorazaWAF")
	return nil
}

func shutdownComponents(cfg *env.ServerConfigHolder, app *Application) error {
	var shutdownErr error

	if app != nil {
		if err := runHooks(app.ShutdownHooks); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown hook failed: %w", err))
		}
		if err := stopBackgroundServices(app.BackgroundServices); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("background service shutdown failed: %w", err))
		}
	}

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

func registerScheduledJobs(jobs []modules.ScheduledJob) error {
	for _, job := range jobs {
		if job.Run == nil {
			continue
		}
		if job.Interval <= 0 {
			return fmt.Errorf("scheduled job %q interval must be greater than 0", job.Name)
		}

		if job.RunOnStart {
			go job.Run()
		}
		scheduler.AddCronJob(job.Interval, job.Run)
		slog.Info("scheduled job registered", "name", job.Name, "interval", job.Interval.String())
	}
	return nil
}

func runHooks(hooks []modules.Hook) error {
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if err := hook(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func startBackgroundServices(services []modules.BackgroundService) error {
	for _, service := range services {
		if service == nil {
			continue
		}
		if err := service.Start(context.Background()); err != nil {
			return fmt.Errorf("start %s failed: %w", service.Name(), err)
		}
		slog.Info("background service started", "name", service.Name())
	}
	return nil
}

func stopBackgroundServices(services []modules.BackgroundService) error {
	var shutdownErr error
	for index := len(services) - 1; index >= 0; index-- {
		service := services[index]
		if service == nil {
			continue
		}
		if err := service.Shutdown(context.Background()); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("stop %s failed: %w", service.Name(), err))
			continue
		}
		slog.Info("background service stopped", "name", service.Name())
	}
	return shutdownErr
}
