package secret

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	SealedKeyLength = nonceLength + KeyLength + secretbox.Overhead
	nonceLength     = 24
)

var ErrInvalidKeySize = errors.New("invalid key size")

type SealedKey struct {
	v [SealedKeyLength]byte
}

func SealedKeyFromBase64(str string) (*SealedKey, error) {
	key := SealedKey{}

	_, err := base64.RawStdEncoding.Strict().Decode(key.v[:], []byte(str))
	if err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	return &key, nil
}

func SealKey(encryptionKey, input *Key) (*SealedKey, error) {
	return sealKey(&encryptionKey.v, input)
}

func SealKeyWithEnclave(enclave *memguard.Enclave, input *Key) (*SealedKey, error) {
	buff, err := enclave.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open the master key enclave: %w", err)
	}

	return sealKey(buff.ByteArray32(), input)
}

func sealKey(encryptionKey *[KeyLength]byte, input *Key) (*SealedKey, error) {
	var nonce [nonceLength]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return nil, fmt.Errorf("failed to generate random numbers: %w", err)
	}

	encrypted := secretbox.Seal(nonce[:], input.Raw(), &nonce, encryptionKey)

	if len(encrypted) != SealedKeyLength {
		return nil, fmt.Errorf("%w: have %d, expected %d", ErrInvalidKeySize, len(encrypted), SealedKeyLength)
	}

	return &SealedKey{v: [SealedKeyLength]byte(encrypted)}, nil
}

func (k *SealedKey) Open(encryptionKey *Key) (*Key, error) {
	return k.open(&encryptionKey.v)
}

func (k *SealedKey) OpenWithEnclave(enclave *memguard.Enclave) (*Key, error) {
	buff, err := enclave.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open the master key enclave: %w", err)
	}

	return k.open(buff.ByteArray32())
}

func (k *SealedKey) open(encryptionKey *[KeyLength]byte) (*Key, error) {
	var decryptNonce [nonceLength]byte
	copy(decryptNonce[:], k.v[:nonceLength])

	decrypted, ok := secretbox.Open(nil, k.v[nonceLength:], &decryptNonce, encryptionKey)
	if !ok {
		return nil, errors.New("failed to open the sealed key")
	}

	if len(decrypted) != KeyLength {
		return nil, fmt.Errorf("%w: have %d, expected %d", ErrInvalidKeySize, len(decrypted), KeyLength)
	}

	return &Key{v: [KeyLength]byte(decrypted)}, nil
}

// String implements the fmt.Stringer interface and returns only the redact hint. This prevents the
// secret value from being printed to std*, logs etc.
func (k *SealedKey) String() string {
	return RedactText
}

// Raw gives you access to the actual secret value stored inside Key.
func (k *SealedKey) Base64() string {
	return base64.RawStdEncoding.Strict().EncodeToString(k.v[:])
}

func (k *SealedKey) Raw() []byte {
	return bytes.Clone(k.v[:])
}

// Equals checks whether s2 has same secret string or not.
func (k *SealedKey) Equals(s2 *SealedKey) bool {
	return k.v == s2.v
}

// MarshalKey implements [encoding.KeyMarshaler]. It marshals redact string into bytes rather than the actual
// secret value.
func (k *SealedKey) MarshalText() ([]byte, error) {
	return []byte(RedactText), nil
}

// UnmarshalKey implements [encoding.KeyUnmarshaler]. It unmarshals b into receiver's new secret value.
// If redact string is present then it is reused otherwise [DefaultRedact] is used.
func (k *SealedKey) UnmarshalText(b []byte) error {
	v := string(b)

	// If the original redact is not nil then use it otherwise fallback to default.
	res, err := SealedKeyFromBase64(v)
	if err != nil {
		return err
	}

	*k = *res

	return nil
}

// MarshalJSON allows Key to be serialized into a JSON string. Only the redact hint is part of the
// the JSON string.
func (k SealedKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, RedactText)), nil
}

// UnmarshalJSON allows a JSON string to be deserialized into a Key value. DefaultRedact is set
// as the redact hint.
func (k *SealedKey) UnmarshalJSON(b []byte) error {
	// Get the new secret value from unmarshalled data.
	var n string
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}

	res, err := SealedKeyFromBase64(n)
	if err != nil {
		return err
	}

	*k = *res

	return nil
}

func (k SealedKey) Value() (driver.Value, error) {
	return k.v[:], nil
}

func (s *SealedKey) Scan(src any) error {
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("expected a []byte")
	}

	s.v = [SealedKeyLength]byte(v)

	return nil
}

func (s SealedKey) LogValue() slog.Value {
	return slog.StringValue(RedactText)
}
