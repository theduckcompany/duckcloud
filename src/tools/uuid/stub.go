package uuid

// Stub implementation of uuid.Provider.
type Stub struct {
	UUID string
}

// New stub method.
func (t *Stub) New() UUID {
	return UUID(t.UUID)
}

// Parse stub method.
//
// This method really parse the input.
func (t *Stub) Parse(s string) (UUID, error) {
	return NewProvider().Parse(s)
}
