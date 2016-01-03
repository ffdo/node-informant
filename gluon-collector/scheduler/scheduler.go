package scheduler

import (
	"time"
)

// ScheduledJob holds the state of a scheduled job like the method to execute
// and the Ticker.
type ScheduledJob struct {
	ticker   *time.Ticker
	quitChan chan interface{}
	method   func()
}

// NewJob creates an new ScheduledJob which executes the given method in the
// specified interval. The bool value fireNow determines if the method should
// executed directly or after the specified duration has passed the first time.
// Please note that the job is executed in a single go routine. This means that
// it is possible for the job to block the loop longer than the interval, which
// causes the method to executed irregularly.
func NewJob(interval time.Duration, job func(), fireNow bool) *ScheduledJob {
	if fireNow {
		go job()
	}
	sJob := &ScheduledJob{
		ticker:   time.NewTicker(interval),
		method:   job,
		quitChan: make(chan interface{}),
	}
	go sJob.loop()
	return sJob
}

// Stop stops a scheduled job immediately. But it will not cancel the
// specified method if it is long running job.
func (s *ScheduledJob) Stop() {
	s.quitChan <- nil
}

func (s *ScheduledJob) loop() {
	for {
		select {
		case <-s.ticker.C:
			s.method()
		case <-s.quitChan:
			s.ticker.Stop()
			return
		}
	}
}
