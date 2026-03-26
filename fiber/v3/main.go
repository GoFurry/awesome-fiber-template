package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/GoFurry/awesome-go-template/fiber/v3/apps/schedule"
	"github.com/GoFurry/awesome-go-template/fiber/v3/common"
	"github.com/GoFurry/awesome-go-template/fiber/v3/common/log"
	cs "github.com/GoFurry/awesome-go-template/fiber/v3/common/service"
	"github.com/GoFurry/awesome-go-template/fiber/v3/middleware"
	"github.com/GoFurry/awesome-go-template/fiber/v3/roof/db"
	"github.com/GoFurry/awesome-go-template/fiber/v3/roof/env"
	"github.com/GoFurry/awesome-go-template/fiber/v3/router"
	corazalite "github.com/GoFurry/coraza-fiber-lite"
	"github.com/gofiber/fiber/v3"
	"github.com/kardianos/service"
)

var errChan = make(chan error)

func main() {
	cfg := env.GetServerConfig()

	exePath, err := os.Executable()
	if err != nil {
		slog.Error("resolve executable path failed", "error", err)
		os.Exit(1)
	}

	appID := cfg.Server.AppID
	if appID == "" {
		appID = common.COMMON_PROJECT_NAME
	}
	appName := cfg.Server.AppName
	if appName == "" {
		appName = appID
	}

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
	prg := &app{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		slog.Error(err.Error())
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err = s.Install()
			if err != nil {
				slog.Error("service install failed", "error", err)
			} else {
				slog.Info("service installed")
			}
			return
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				slog.Error("service uninstall failed", "error", err)
			} else {
				slog.Info("service uninstalled")
			}
			return
		case "version":
			slog.Info(appName + " " + cfg.Server.AppVersion)
			return
		case "help":
			slog.Info(common.COMMON_PROJECT_HELP)
			return
		}
		return
	}

	debug.SetGCPercent(cfg.Server.GCPercent)
	debug.SetMemoryLimit(int64(cfg.Server.MemoryLimit << 30))

	InitOnStart()
	err = s.Run()
	if err != nil {
		slog.Error(err.Error())
	}
}

func InitOnStart() {
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

	middleware.InitPrometheus(middleware.FiberPromConf{
		SkipPaths:         []string{},
		IgnoreStatusCodes: []int{},
	})

	if cfg.Waf.Enabled {
		corazalite.InitMetrics()
		corazalite.InitGlobalWAFWithCfg(corazalite.CorazaCfg{
			DirectivesFile:     cfg.Waf.ConfPath,
			RequestBodyAccess:  true,
			ResponseBodyAccess: false,
		})
		corazalite.InitWAFBlockMessage("Request blocked by CorazaLite WAF")
	}

	if err := db.InitDatabaseOnStart(); err != nil {
		slog.Error("database init failed", "error", err)
		os.Exit(1)
	}

	if err := cs.InitRedisOnStart(); err != nil {
		slog.Warn("redis init skipped", "error", err)
	}

	if cfg.Schedule.Enabled {
		cs.InitTimeWheelOnStart()
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
