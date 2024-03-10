package spaces

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const BootstrapSpaceName = "Everyone"

type Space struct {
	createdAt time.Time
	id        uuid.UUID
	name      string
	createdBy uuid.UUID
	owners    Owners
}

func (f Space) ID() uuid.UUID        { return f.id }
func (f Space) Name() string         { return f.name }
func (f Space) Owners() Owners       { return f.owners }
func (f Space) CreatedAt() time.Time { return f.createdAt }
func (f Space) CreatedBy() uuid.UUID { return f.createdBy }

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

	*t = Owners{}

	if len(s) == 0 {
		return nil
	}

	rawIDs := strings.Split(s, ",")

	for _, id := range rawIDs {
		*t = append(*t, uuid.UUID(id))
	}

	return nil
}

type CreateCmd struct {
	User   *users.User
	Name   string
	Owners []uuid.UUID
}

// Validate the fields.
func (t CreateCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.User, v.Required),
		v.Field(&t.Name, v.Required, v.Length(1, 30)),
		v.Field(&t.Owners, v.Each(is.UUIDv4)),
	)
}

type AddOwnerCmd struct {
	User    *users.User
	Owner   *users.User
	SpaceID uuid.UUID
}

// Validate the fields.
func (t AddOwnerCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.User, v.Required),
		v.Field(&t.Owner, v.Required),
		v.Field(&t.SpaceID, v.Required, is.UUIDv4),
	)
}

type RemoveOwnerCmd struct {
	User    *users.User
	Owner   *users.User
	SpaceID uuid.UUID
}

// Validate the fields.
func (t RemoveOwnerCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.User, v.Required),
		v.Field(&t.Owner, v.Required),
		v.Field(&t.SpaceID, v.Required, is.UUIDv4),
	)
}
