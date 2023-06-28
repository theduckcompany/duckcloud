package tools

import (
	"testing"

	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/jwt"
	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/neilotoole/slogt"
)

func NewMockTools(t *testing.T) Tools {
	log := slogt.New(t)

	return Tools{
		Clock:     clock.NewMockClock(t),
		UUID:      uuid.NewMockProvider(t),
		Log:       log,
		ResWriter: response.New(log),
		JWT:       jwt.NewMockParser(t),
	}
}
