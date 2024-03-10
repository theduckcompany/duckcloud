package sqlstorage

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type SQLTime time.Time

func (t SQLTime) Value() (driver.Value, error) {
	tt := time.Time(t)

	resStr, err := tt.UTC().MarshalText()
	if err != nil {
		return nil, err
	}

	return string(resStr), nil
}

func (t *SQLTime) Scan(value any) error {
	tt := (*time.Time)(t)

	switch v := value.(type) {
	case string:
		err := (*tt).UnmarshalText([]byte(v))
		(*tt) = tt.UTC()
		return err
	default:
		return fmt.Errorf("unsuported type: %T", v)
	}
}

func (t SQLTime) Time() time.Time {
	return time.Time(t)
}
