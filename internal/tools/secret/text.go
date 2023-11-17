// Package secret provides types to guard your secret values from leaking into logs, std* etc.
//
// The objective is to disallow writing/serializing of secret values to std*, logs, JSON string
// etc. but provide access to the secret when requested explicitly.
package secret

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log/slog"
)

// RedactText is used by default if no other redact hint is given.
const RedactText string = "*****"

var Empty = NewText("")

// Text provides a way to safely store your secret value and a corresponding redact hint. This
// redact hint what is used in operations like printing and serializing. The default
// value of Text is usable.
type Text struct {
	// v is the actual secret values.
	v string
}

// NewText creates a new Text instance with s as the secret value. Multiple option functions can
// be passed to alter default behavior.
func NewText(s string) Text {
	sec := Text{
		v: s,
	}

	return sec
}

// String implements the fmt.Stringer interface and returns only the redact hint. This prevents the
// secret value from being printed to std*, logs etc.
func (s Text) String() string {
	return RedactText
}

// Raw gives you access to the actual secret value stored inside Text.
func (s Text) Raw() string {
	return s.v
}

// MarshalText implements [encoding.TextMarshaler]. It marshals redact string into bytes rather than the actual
// secret value.
func (s Text) MarshalText() ([]byte, error) {
	return []byte(RedactText), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]. It unmarshals b into receiver's new secret value.
// If redact string is present then it is reused otherwise [RedactText] is used.
func (s *Text) UnmarshalText(b []byte) error {
	v := string(b)

	// If the original redact is not nil then use it otherwise fallback to default.
	*s = NewText(v)
	return nil
}

// MarshalJSON allows Text to be serialized into a JSON string. Only the redact hint is part of the
// the JSON string.
func (s Text) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, RedactText)), nil
}

// UnmarshalJSON allows a JSON string to be deserialized into a Text value. RedactText is set
// as the redact hint.
func (s *Text) UnmarshalJSON(b []byte) error {
	// Get the new secret value from unmarshalled data.
	var n string
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}

	*s = NewText(n)

	return nil
}

// Equals checks whether s2 has same secret string or not.
func (s *Text) Equals(s2 Text) bool {
	return s.v == s2.v
}

func (s Text) Value() (driver.Value, error) {
	return s.v, nil
}

func (s *Text) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("not a string")
	}

	s.v = str

	return nil
}

func (s Text) LogValue() slog.Value {
	return slog.StringValue(RedactText)
}
