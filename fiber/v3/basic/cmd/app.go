package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/basic/config"
	cache "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/cache"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/db"
	log "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/logging"
	scheduler "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/scheduler"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/schedule"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/transport/http/middleware"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/transport/http/router"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/pkg/common"
	corazalite "github.com/GoFurry/coraza-fiber-lite"
	"github.com/gofiber/fiber/v3"
	"github.com/kardianos/service"
)

var errChan = make(chan error)

func runService() error {
	cfg := env.GetServerConfig()
	svc, err := newService()
	if err != nil {
		return err
	}

	debug.SetGCPercent(cfg.Server.GCPercent)
	debug.SetMemoryLimit(int64(cfg.Server.MemoryLimit << 30))

	initOnStart()
	return svc.Run()
}

func newService() (service.Service, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path failed: %w", err)
	}

	appID, appName := appIdentity()
	svcConfig := &service.Config{
		Name:        appID,
		DisplayName: appName,
		Description: appName,
		Option: service.KeyValue{
			"SystemdScript": `[Unit]
Description=` + appName + `
After=network.target
Requires=network.target

[Service]
Type=simple
WorkingDirectory=` + filepath.Dir(exePath) + `/
ExecStart=` + exePath + `
Restart=always
RestartSec=30
LogOutput=true
LogDirectory=/var/log/` + appID + `
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target`,
		},
	}

	return service.New(&app{}, svcConfig)
}

func appIdentity() (string, string) {
	cfg := env.GetServerConfig()
	appID := cfg.Server.AppID
	if appID == "" {
		appID = common.COMMON_PROJECT_NAME
	}

	appName := cfg.Server.AppName
	if appName == "" {
		appName = appID
	}
	return appID, appName
}

func initOnStart() {
	cfg := env.GetServerConfig()

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

	err := log.InitLogger(logCfg)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if cfg.Prometheus.Enabled {
		middleware.InitPrometheus(middleware.FiberPromConf{
			SkipPaths:         []string{},
			IgnoreStatusCodes: []int{},
		})
	}

	if cfg.Waf.Enabled {
		corazalite.InitMetrics()
		corazalite.InitGlobalWAFWithCfg(corazalite.CorazaCfg{
			DirectivesFile:     cfg.Waf.ConfPath,
			RequestBodyAccess:  true,
			ResponseBodyAccess: false,
		})
		corazalite.InitWAFBlockMessage("Request blocked by CorazaLite WAF")
	}

	if cfg.DataBase.Enabled {
		if err := db.InitDatabaseOnStart(); err != nil {
			slog.Error("database init failed", "error", err)
			os.Exit(1)
		}
	}

	if cfg.Redis.Enabled {
		cache.InitRedisOnStart()
	}

	if cfg.Schedule.Enabled {
		scheduler.InitTimeWheelOnStart()
		schedule.InitScheduleOnStart()
	}
}

type app struct{}

func (a *app) Start(s service.Service) error {
	go a.run()
	return nil
}

func (a *app) run() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		cfg := env.GetServerConfig()
		app := router.Router.Init()
		addr := cfg.Server.IPAddress + ":" + cfg.Server.Port

		if err := app.Listen(addr, fiber.ListenConfig{
			TLSConfig:         nil,
			EnablePrefork:     cfg.Server.EnablePrefork,
			ListenerNetwork:   cfg.Server.Network,
			EnablePrintRoutes: cfg.Server.Mode == "debug",
		}); err != nil {
			errChan <- err
		}
	}()

	if err := <-errChan; err != nil {
		slog.Error(err.Error())
	}
}

func (a *app) Stop(s service.Service) error {
	return nil
}
