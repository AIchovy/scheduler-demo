package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/mosesyou/scheduler-demo/pkg/framework"
	"github.com/mosesyou/scheduler-demo/pkg/job"
)

// task1 interval job
func task1() error {
	log.Info("task1 running.....")

	return nil
}

// task2 runOnce job
func task2() error {
	log.Info("task2 running.....")

	return nil
}

// task3 error job
func task3() error {
	log.Info("task3 running.....")

	return fmt.Errorf("mock job faild")
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	job1 := job.NewJob("task1").AddFunc(task1).
		SetRunOnce(false).SetInterval(10 * time.Second)

	job2 := job.NewJob("task2").AddFunc(task2).
		SetRunOnce(true)

	job3 := job.NewJob("task3").AddFunc(task3).
		SetRunOnce(true).SetRetry(3)

	scheduler := framework.NewScheduler()

	_ = scheduler.ScheduleJob(job1)
	_ = scheduler.ScheduleJob(job2)
	_ = scheduler.ScheduleJob(job3)

	scheduler.Start()
}
