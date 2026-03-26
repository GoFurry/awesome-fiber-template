package schedule

import (
	"time"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/config"
	modules "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/schedule/task"
)

func NewBundle() (modules.Bundle, error) {
	return modules.Bundle{
		ScheduledJobs: []modules.ScheduledJob{
			{
				Name:       "schedule.metrics_cache_refresh",
				Interval:   10 * time.Minute,
				RunOnStart: true,
				Run:        ScheduleByTenMinutes,
			},
			{
				Name:       "schedule.hourly_housekeeping",
				Interval:   1 * time.Hour,
				RunOnStart: true,
				Run:        ScheduleByOneHour,
			},
		},
	}, nil
}

func ScheduleByTenMinutes() {
	cfg := env.GetServerConfig()

	if cfg.Prometheus.Enabled {
		task.UpdateMetricsCache()
	}
}

func ScheduleByOneHour() {
}
