package clock

import "time"

// Stub is a stub implementation of Clock
type Stub struct {
	Time time.Time
}

// Now stub method.
func (t *Stub) Now() time.Time {
	return t.Time
}
