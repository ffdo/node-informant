package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobScheduling(t *testing.T) {
	assert := assert.New(t)

	executed := false

	job := NewJob(time.Second*1, func() {
		executed = true
	})
	time.Sleep(time.Second * 2)
	job.Stop()
	assert.True(executed)
}
