package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

func TestCron(t *testing.T) {
	t.Run("Start an async job and stop it with Stop", func(t *testing.T) {
		tools := tools.NewMock(t)
		cronRunner := NewMockCronRunner(t)

		runner := New("test-1", time.Second, tools, cronRunner)

		// Start the async job. The first call is done 1s after the call to Start
		go func(cron *Cron) {
			cron.RunLoop()
		}(runner)

		// Stop will interrupt the job before the second.
		runner.Stop()
	})

	t.Run("Stop interrupt the running job", func(t *testing.T) {
		tools := tools.NewMock(t)
		cronRunner := NewMockCronRunner(t)

		runner := New("test-1", time.Millisecond, tools, cronRunner)

		go runner.RunLoop()

		// First loop to fetch the deleted inodes. Make it take more than a 1s.
		cronRunner.On("Run", mock.Anything).WaitUntil(time.After(time.Minute)).Return(nil).Once()

		// Wait some time in order to be just to have the job running and waiting for the end of "GetAllDeleted".
		time.Sleep(20 * time.Millisecond)

		// Stop will interrupt the job before the second.
		start := time.Now()
		runner.Stop()
		assert.WithinDuration(t, time.Now(), start, 10*time.Millisecond)
	})
}
