package server

import (
	"log"
	"time"

	"github.com/go-co-op/gocron"
)

type scheduler struct {
	*gocron.Scheduler
}

func startScheduler() (s *scheduler) {
	log.Printf("[INFO] Starting scheduler")

	gs := gocron.NewScheduler(time.UTC)

	scheduler := &scheduler{
		Scheduler: gs,
	}

	gs.StartAsync()

	return scheduler
}

func (s *scheduler) scheduleAt(t time.Time, topic string, job func()) {
	_, err := s.Scheduler.Every(1).Millisecond().StartAt(t).Tag(topic).LimitRunsTo(1).Do(job)
	if err != nil {
		log.Printf("[ERROR] Failed to schedule job: %v", err)
	}
}

// func (s *scheduler) scheduleCron(cron string, job func()) {
// 	s.Scheduler.Cron(cron).Do(job)
// }

// func (s *scheduler) scheduleImmediate(job func()) {
// 	s.Scheduler.Every(1).Millisecond().LimitRunsTo(1).Do(job)
// }
