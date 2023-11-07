package password

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type Argon2IDPassword struct{}

func NewArgon2IDPassword() *Argon2IDPassword {
	return &Argon2IDPassword{}
}

func (p *Argon2IDPassword) Encrypt(ctx context.Context, password string) (string, error) {
	// Those values have been taken from https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
	params := &params{
		saltLength:  16,
		iterations:  3,
		parallelism: 1,
		memory:      12 * 1024, // 12MB
		keyLength:   32,
	}

	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(params.saltLength)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.memory, params.iterations, params.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func (p *Argon2IDPassword) Compare(ctx context.Context, hashStr, password string) (bool, error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	params, salt, hash, err := decodeHash(hashStr)
	if err != nil {
		return false, fmt.Errorf("failed to decode the hash: %w", err)
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
