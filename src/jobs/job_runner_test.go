package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/tools"
)

type jobMock struct {
	mock.Mock
}

func newJobMock(t *testing.T) *jobMock {
	mock := new(jobMock)
	mock.Test(t)

	return mock
}

func (m *jobMock) Run(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestJobRunner(t *testing.T) {
	t.Run("Start an async job and stop it with Stop", func(t *testing.T) {
		tools := tools.NewMock(t)
		jobMock := newJobMock(t)

		runner := NewJobRunner(jobMock, time.Second, tools)

		// Start the async job. The first call is done 1s after the call to Start
		runner.Start()

		// Stop will interrupt the job before the second.
		runner.Stop()
	})

	t.Run("Stop interrupt the running job", func(t *testing.T) {
		tools := tools.NewMock(t)
		jobMock := newJobMock(t)

		runner := NewJobRunner(jobMock, time.Millisecond, tools)

		runner.Start()

		// First loop to fetch the deleted inodes. Make it take more than a 1s.
		jobMock.On("Run", mock.Anything).WaitUntil(time.After(time.Minute)).Return(nil).Once()

		// Wait some time in order to be just to have the job running and waiting for the end of "GetAllDeleted".
		time.Sleep(20 * time.Millisecond)

		// Stop will interrupt the job before the second.
		start := time.Now()
		runner.Stop()
		assert.WithinDuration(t, time.Now(), start, 10*time.Millisecond)
	})
}
