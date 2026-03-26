package bootstrap

import (
	modules "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules"
	schedulemodule "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/schedule"
	usermodule "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user"
)

type Application struct {
	RouteModules       []modules.RouteModule
	DatabaseModels     []any
	Migrations         []modules.Migration
	StartupHooks       []modules.Hook
	ShutdownHooks      []modules.Hook
	ScheduledJobs      []modules.ScheduledJob
	BackgroundServices []modules.BackgroundService
}

func buildApplication() (*Application, error) {
	bundle, err := modules.Collect(
		schedulemodule.NewBundle,
		usermodule.NewBundle,
	)
	if err != nil {
		return nil, err
	}

	return &Application{
		RouteModules:       bundle.RouteModules,
		DatabaseModels:     bundle.DatabaseModels,
		Migrations:         bundle.Migrations,
		StartupHooks:       bundle.StartupHooks,
		ShutdownHooks:      bundle.ShutdownHooks,
		ScheduledJobs:      bundle.ScheduledJobs,
		BackgroundServices: bundle.BackgroundServices,
	}, nil
}
