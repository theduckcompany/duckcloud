package uuid

import "github.com/google/uuid"

// Default implementation of UUIDProvider
type Default struct{}

// NewProvider return a new Default uuid provider.
func NewProvider() *Default {
	return &Default{}
}

// New implementation of uuid.Provider
func (t Default) New() UUID {
	return UUID(uuid.Must(uuid.NewRandom()).String())
}

// Parse implementation of uuid.Provider
func (t Default) Parse(s string) (UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return UUID(""), err
	}

	return UUID(u.String()), nil
}
