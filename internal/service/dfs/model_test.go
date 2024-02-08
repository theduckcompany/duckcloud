package dfs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestInodeGetter(t *testing.T) {
	assert.Equal(t, ExampleAliceRoot.ID(), uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"))
	assert.Equal(t, "", ExampleAliceRoot.Name())
	assert.Nil(t, ExampleAliceRoot.Parent())
	assert.Nil(t, ExampleAliceRoot.FileID())
	assert.Equal(t, ExampleAliceRoot.CreatedAt(), now)
	assert.Equal(t, ExampleAliceRoot.CreatedBy(), users.ExampleAlice.ID())
	assert.Equal(t, ExampleAliceRoot.LastModifiedAt(), now2)
	assert.True(t, ExampleAliceRoot.IsDir())

	assert.Equal(t, uint64(42), ExampleAliceFile.Size())
	assert.False(t, ExampleAliceFile.IsDir())
	assert.Equal(t, ExampleAliceFile.FileID(), ptr.To(uuid.UUID("abf05a02-8af9-4184-a46d-847f7d951c6b")))
	assert.Equal(t, ExampleAliceFile.Parent(), ExampleAliceFile.parent)
	assert.Equal(t, ExampleAliceFile.SpaceID(), ExampleAliceFile.spaceID)
}

func Test_Inodes_Commands(t *testing.T) {
	t.Run("CreateRootDirCmd", func(t *testing.T) {
		cmd := CreateRootDirCmd{
			CreatedBy: nil,
			Space:     &spaces.ExampleAlicePersonalSpace,
		}

		err := cmd.Validate()
		require.EqualError(t, err, "CreatedBy: cannot be blank.")
	})
}

func Test_PathCmd_Equal(t *testing.T) {
	tests := []struct {
		A        *PathCmd
		B        *PathCmd
		Expected bool
	}{
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			Expected: true,
		},
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"),
			Expected: false,
		},
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			Expected: false,
		},
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/"),
			B:        NewPathCmd(&spaces.ExampleBobPersonalSpace, "/foo/"),
			Expected: false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			res := test.A.Equal(*test.B)
			if res != test.Expected {
				t.Fatalf("%q .Equal %q -> have %v, expected: %v", test.A, test.B, res, test.Expected)
			}
		})
	}
}

func Test_PathCmd_Contains(t *testing.T) {
	require.Implements(t, (*fmt.Stringer)(nil), new(PathCmd))

	tests := []struct {
		A        *PathCmd
		B        *PathCmd
		Expected bool
	}{
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			Expected: true,
		},
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"),
			Expected: true,
		},
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/bar"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			Expected: false,
		},
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo/"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			Expected: true,
		},
		{
			A:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "//foo"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo//"),
			Expected: true,
		},
		{
			A:        NewPathCmd(&spaces.ExampleBobPersonalSpace, "/foo"),
			B:        NewPathCmd(&spaces.ExampleAlicePersonalSpace, "/foo"),
			Expected: false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			res := test.A.Contains(*test.B)
			if res != test.Expected {
				t.Fatalf("%q .Contains %q -> have %v, expected: %v", test.A, test.B, res, test.Expected)
			}
		})
	}
}

func Test_PathCmd_String(t *testing.T) {
	assert.Equal(t, NewPathCmd(&spaces.ExampleBobPersonalSpace, "/foo").String(), fmt.Sprintf("%s:/foo", spaces.ExampleBobPersonalSpace.ID()))
}
