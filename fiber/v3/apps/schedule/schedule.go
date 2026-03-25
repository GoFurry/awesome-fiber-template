package schedule

import (
	"fmt"
	"time"

	"github.com/GoFurry/awesome-go-template/fiber/v3/apps/schedule/task"
	"github.com/GoFurry/awesome-go-template/fiber/v3/common/log"
	cs "github.com/GoFurry/awesome-go-template/fiber/v3/common/service"
)

// 初始化
func InitScheduleOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("receive InitScheduleOnStart recover: %v", err))
		}
	}()
	log.Info("Schedule 模块初始化开始...")

	//初始化后执行一次 Schedule
	go ScheduleByTenMinutes()
	go ScheduleByOneHour()
	// 定时任务执行 Schedule
	cs.AddCronJob(10*time.Minute, ScheduleByTenMinutes)
	cs.AddCronJob(1*time.Hour, ScheduleByOneHour)

	log.Info("Schedule 模块初始化结束...")
}

// 十分钟任务表
func ScheduleByTenMinutes() {
	task.UpdateMetricsCache()
}

// 一小时任务表
func ScheduleByOneHour() {

}
