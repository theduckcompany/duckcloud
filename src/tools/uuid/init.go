package uuid

// UUID custome type
type UUID string

type Service interface {
	// New create a new UUID V4
	New() UUID
	Parse(string) (UUID, error)
}
