package uuid

// UUID custome type
type UUID string

//go:generate mockery --name Service
type Service interface {
	// New create a new UUID V4
	New() UUID
	Parse(string) (UUID, error)
}
