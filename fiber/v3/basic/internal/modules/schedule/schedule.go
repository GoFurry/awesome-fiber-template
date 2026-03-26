package schedule

import (
	"fmt"
	"time"

	log "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/logging"
	scheduler "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/scheduler"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/schedule/task"
)

func InitScheduleOnStart() (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Error(fmt.Sprintf("receive InitScheduleOnStart recover: %v", recovered))
			err = fmt.Errorf("init schedule panic: %v", recovered)
		}
	}()

	log.Info("schedule module initialization started")

	go ScheduleByTenMinutes()
	go ScheduleByOneHour()

	scheduler.AddCronJob(10*time.Minute, ScheduleByTenMinutes)
	scheduler.AddCronJob(1*time.Hour, ScheduleByOneHour)

	log.Info("schedule module initialization finished")
	return nil
}

func ScheduleByTenMinutes() {
	task.UpdateMetricsCache()
}

func ScheduleByOneHour() {
}
