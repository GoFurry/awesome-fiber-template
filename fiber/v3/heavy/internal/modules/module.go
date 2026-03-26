package modules

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type RouteModule interface {
	Name() string
	RegisterRoutes(root fiber.Router)
}

type Hook func(ctx context.Context) error

type BackgroundService interface {
	Name() string
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type ScheduledJob struct {
	Name       string
	Interval   time.Duration
	RunOnStart bool
	Run        func()
}

type Migration interface {
	Name() string
	Up(tx *gorm.DB) error
}

type Bundle struct {
	RouteModules       []RouteModule
	DatabaseModels     []any
	Migrations         []Migration
	StartupHooks       []Hook
	ShutdownHooks      []Hook
	ScheduledJobs      []ScheduledJob
	BackgroundServices []BackgroundService
}

type Factory func() (Bundle, error)

func Collect(factories ...Factory) (Bundle, error) {
	result := Bundle{}
	for _, factory := range factories {
		if factory == nil {
			continue
		}

		bundle, err := factory()
		if err != nil {
			return Bundle{}, err
		}

		result.RouteModules = append(result.RouteModules, bundle.RouteModules...)
		result.DatabaseModels = append(result.DatabaseModels, bundle.DatabaseModels...)
		result.Migrations = append(result.Migrations, bundle.Migrations...)
		result.StartupHooks = append(result.StartupHooks, bundle.StartupHooks...)
		result.ShutdownHooks = append(result.ShutdownHooks, bundle.ShutdownHooks...)
		result.ScheduledJobs = append(result.ScheduledJobs, bundle.ScheduledJobs...)
		result.BackgroundServices = append(result.BackgroundServices, bundle.BackgroundServices...)
	}

	return result, nil
}
