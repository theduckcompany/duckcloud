package clock

import "time"

// Default is a Clock implementation base ton time.Now()
type Default struct{}

// New create a new Clock.
func New() *Default {
	return &Default{}
}

// Now return the time for the exact moment.
func (t *Default) Now() time.Time {
	return time.Now().UTC()
}
