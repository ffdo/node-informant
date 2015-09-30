package scheduler

import (
	"time"
)

type ScheduledJob struct {
	ticker   *time.Ticker
	quitChan chan bool
	method   func()
}

func NewJob(interval time.Duration, job func()) *ScheduledJob {
	sJob := &ScheduledJob{
		ticker:   time.NewTicker(interval),
		method:   job,
		quitChan: make(chan bool),
	}
	go func() {
		sJob.loop()
	}()
	return sJob
}

func (s *ScheduledJob) Stop() {
	s.quitChan <- true
}

func (s *ScheduledJob) loop() {
	stopped := false
	for !stopped {
		select {
		case <-s.ticker.C:
			s.method()
		case quit := <-s.quitChan:
			s.ticker.Stop()
			stopped = quit
		}
	}
}
