package usercreate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestUserCreateJob(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		usersMock := users.NewMockService(t)
		inodesMock := inodes.NewMockService(t)
		tools := tools.NewMock(t)

		job := NewJob(usersMock, inodesMock, tools)

		usersMock.On("GetAllWithStatus", mock.Anything, "initializing", &storage.PaginateCmd{Limit: batchSize}).
			Return([]users.User{users.ExampleInitializingAlice}, nil).Once()
		inodesMock.On("CreateRootDir", mock.Anything, users.ExampleAlice.ID()).Return(&inodes.ExampleAliceRoot, nil).Once()
		usersMock.On("SaveBootstrapInfos", mock.Anything, users.ExampleAlice.ID(), &inodes.ExampleAliceRoot).Return(&users.ExampleAlice, nil).Once()

		err := job.Run(ctx)
		assert.NoError(t, err)
	})
}
