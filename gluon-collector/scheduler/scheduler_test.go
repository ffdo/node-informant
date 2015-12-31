package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobScheduling(t *testing.T) {
	assert := assert.New(t)

	executed := false
	now := time.Now()
	executionCount := 0
	job := NewJob(time.Second*1, func() {
		then := time.Now()
		diff := then.Unix() - now.Unix()
		assert.False(diff < 1, "Less than a second has passed. Diff %d", diff)
		assert.False(diff >= 2, "Two or more seconds have passed. Diff %d", diff)
		executed = true
		now = time.Now()
		executionCount = executionCount + 1
	}, false)
	time.Sleep(time.Second * 3)
	job.Stop()
	assert.True(executed, "The scheduler has NOT been executed")
	assert.True(2 <= executionCount, "The scheduler has only been executed twice or less time")
}
