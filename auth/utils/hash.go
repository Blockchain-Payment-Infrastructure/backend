package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
	SaltLength  uint32
}

// Default Parameters
var defaultParams = Argon2Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	KeyLength:   32,
	SaltLength:  16,
}

func HashPassword(password string) (string, error) {
	// generate random salt, unique for each password
	salt := make([]byte, defaultParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// hash the password with salt and default parameters
	hash := argon2.IDKey([]byte(password), salt, defaultParams.Iterations, defaultParams.Memory, defaultParams.Parallelism, defaultParams.KeyLength)

	// encode salt and hash to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// store them in the encoded format recognised by argon2
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, defaultParams.Memory, defaultParams.Iterations, defaultParams.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

var (
	ErrInvalidHash         = errors.New("The encoded hash is not in the correct format")
	ErrIncompatibleParams  = errors.New("Incompatible hash parameters")
	ErrIncompatibleVersion = errors.New("Incompatible version of argon2")
)

func VerifyPassword(encodedHash, password string) (bool, error) {
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// hash the password we got from the user
	otherHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	// if both hashes don't match
	if subtle.ConstantTimeCompare(hash, otherHash) == 0 {
		return false, nil
	}

	// if both hashes match
	return true, nil
}

func decodeHash(encodedHash string) (p *Argon2Params, salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("Invalid hash format")
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &Argon2Params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	if defaultParams.Iterations != p.Iterations && defaultParams.Memory != p.Memory && defaultParams.Parallelism != p.Parallelism {
		return nil, nil, nil, ErrIncompatibleParams
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
