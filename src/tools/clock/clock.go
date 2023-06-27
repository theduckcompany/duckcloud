package clock

import "time"

// Clock is used to give time.
type Clock interface {
	Now() time.Time
}
