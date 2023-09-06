package folders

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type Folder struct {
	id             uuid.UUID
	name           string
	isPublic       bool
	owners         Owners
	size           uint64
	rootFS         uuid.UUID
	createdAt      time.Time
	lastModifiedAt time.Time
}

func (f *Folder) ID() uuid.UUID             { return f.id }
func (f *Folder) Name() string              { return f.name }
func (f *Folder) IsPublic() bool            { return f.isPublic }
func (f *Folder) Owners() Owners            { return f.owners }
func (f *Folder) RootFS() uuid.UUID         { return f.rootFS }
func (f *Folder) Size() uint64              { return f.size }
func (f *Folder) CreatedAt() time.Time      { return f.createdAt }
func (f *Folder) LastModifiedAt() time.Time { return f.lastModifiedAt }

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

type CreatePersonalFolderCmd struct {
	Name  string
	Owner uuid.UUID
}

// Validate the fields.
func (t CreatePersonalFolderCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Name, v.Required, v.Length(1, 30)),
		v.Field(&t.Owner, v.Required, is.UUIDv4),
	)
}
