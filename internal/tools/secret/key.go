package secret

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
)

const DefaultKeyLength = 32

type Key struct {
	// v is the actual secret values.
	v [DefaultKeyLength]byte
}

func NewKey() (*Key, error) {
	var key Key

	_, err := rand.Read(key.v[:])
	if err != nil {
		return nil, fmt.Errorf("failed to generate randomness: %w", err)
	}

	return &key, nil
}

func FromBase64(str string) (*Key, error) {
	key := Key{}

	_, err := base64.RawStdEncoding.Strict().Decode(key.v[:], []byte(str))
	if err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	return &key, nil
}

// String implements the fmt.Stringer interface and returns only the redact hint. This prevents the
// secret value from being printed to std*, logs etc.
func (s Key) String() string {
	return DefaultRedact
}

// Raw gives you access to the actual secret value stored inside Key.
func (s Key) Base64() string {
	return base64.RawStdEncoding.Strict().EncodeToString(s.v[:])
}

// MarshalKey implements [encoding.KeyMarshaler]. It marshals redact string into bytes rather than the actual
// secret value.
func (s Key) MarshalText() ([]byte, error) {
	return []byte(DefaultRedact), nil
}

// UnmarshalKey implements [encoding.KeyUnmarshaler]. It unmarshals b into receiver's new secret value.
// If redact string is present then it is reused otherwise [DefaultRedact] is used.
func (s *Key) UnmarshalText(b []byte) error {
	v := string(b)

	// If the original redact is not nil then use it otherwise fallback to default.
	res, err := FromBase64(v)
	if err != nil {
		return err
	}

	*s = *res

	return nil
}

// MarshalJSON allows Key to be serialized into a JSON string. Only the redact hint is part of the
// the JSON string.
func (s Key) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, DefaultRedact)), nil
}

// UnmarshalJSON allows a JSON string to be deserialized into a Key value. DefaultRedact is set
// as the redact hint.
func (s *Key) UnmarshalJSON(b []byte) error {
	// Get the new secret value from unmarshalled data.
	var n string
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}

	res, err := FromBase64(n)
	if err != nil {
		return err
	}

	*s = *res

	return nil
}

// Equals checks whether s2 has same secret string or not.
func (s *Key) Equals(s2 *Key) bool {
	return s.v == s2.v
}

func (s Key) Value() (driver.Value, error) {
	return s.v[:], nil
}

func (s *Key) Scan(src any) error {
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected a []byte")
	}

	s.v = [DefaultKeyLength]byte(v)

	return nil
}

func (s Key) LogValue() slog.Value {
	return slog.StringValue(DefaultRedact)
}
