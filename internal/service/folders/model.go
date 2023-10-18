package folders

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type Folder struct {
	id        uuid.UUID
	name      string
	isPublic  bool
	owners    Owners
	rootFS    uuid.UUID
	createdAt time.Time
}

func (f *Folder) ID() uuid.UUID        { return f.id }
func (f *Folder) Name() string         { return f.name }
func (f *Folder) IsPublic() bool       { return f.isPublic }
func (f *Folder) Owners() Owners       { return f.owners }
func (f *Folder) RootFS() uuid.UUID    { return f.rootFS }
func (f *Folder) CreatedAt() time.Time { return f.createdAt }

type Owners []uuid.UUID

func (t Owners) String() string {
	rawIDs := make([]string, len(t))

	for i, id := range t {
		rawIDs[i] = string(id)
	}

	return strings.Join(rawIDs, ",")
}

func (t Owners) Value() (driver.Value, error) { return t.String(), nil }
func (t *Owners) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("not a string")
	}

	rawIDs := strings.Split(s, ",")

	for _, id := range rawIDs {
		*t = append(*t, uuid.UUID(id))
	}

	return nil
}

type CreateCmd struct {
	Name   string
	Owners []uuid.UUID
	RootFS uuid.UUID
}

// Validate the fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Name, v.Required, v.Length(1, 30)),
		v.Field(&t.Owners, v.Required, v.Each(is.UUIDv4)),
		v.Field(&t.RootFS, v.Required, is.UUIDv4),
	)
}
