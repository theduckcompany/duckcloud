package clock

import "time"

// Clock is used to give time.
//
//go:generate mockery --name Clock
type Clock interface {
	Now() time.Time
}
