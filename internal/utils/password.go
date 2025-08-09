package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrorShortPassword                = errors.New("password length is less than 8 characters")
	ErrorPasswordTooLong              = errors.New("password length is greater than 64 characters")
	ErrorInvalidCharactersInPassword  = errors.New("alpha numeric and special characters are allowed in the password field")
	ErrorNotEnoughSpecialCharacters   = errors.New("not enough special characters in the password")
	ErrorNotEnoughUpperCaseCharacters = errors.New("not enough upper case characters in the password")
	ErrorNotEnoughLowerCaseCharacters = errors.New("not enough lower case characters in the password")
	ErrorNotEnoughDigits              = errors.New("not enough digits in the password")
)

const (
	ASCIIUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ASCIILower = "abcdefghijklmnopqrstuvwxyz"
	Digits     = "0123456789"
	Special    = "!@#$%^&*()-_=+[]{};:,.<>?/|\\`~"
)

// Validate the password according the password requirements
func ValidatePassword(password string) (bool, error) {
	if len(password) < 8 {
		return false, ErrorShortPassword
	}

	if len(password) > 64 {
		return false, ErrorPasswordTooLong
	}

	specialCount, upperCount, lowerCount, digitCount := 0, 0, 0, 0
	for _, ch := range password {
		// Reject non-ASCII characters
		if ch > 127 {
			return false, ErrorInvalidCharactersInPassword
		}
		// Reject control characters
		if ch < 32 || ch == 127 {
			return false, ErrorInvalidCharactersInPassword
		}

		switch {
		case strings.ContainsRune(Special, ch):
			specialCount++
		case strings.ContainsRune(ASCIIUpper, ch):
			upperCount++
		case strings.ContainsRune(ASCIILower, ch):
			lowerCount++
		case strings.ContainsRune(Digits, ch):
			digitCount++
		}
	}

	if specialCount < 1 {
		return false, ErrorNotEnoughSpecialCharacters
	}
	if upperCount < 1 {
		return false, ErrorNotEnoughUpperCaseCharacters
	}
	if lowerCount < 1 {
		return false, ErrorNotEnoughLowerCaseCharacters
	}
	if digitCount < 1 {
		return false, ErrorNotEnoughDigits
	}

	return true, nil
}

/* Password Hashing Utilities */

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var argon2Params = &params{
	memory:      64 * 1024,
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

func HashPassword(password string) string {
	salt := make([]byte, 0, 16)
	_, _ = rand.Read(salt) // practically doesn't return any error
	hash := argon2.IDKey([]byte(password), salt, argon2Params.iterations, argon2Params.memory, argon2Params.parallelism, argon2Params.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return encodeHash(b64Salt, b64Hash)
}

func encodeHash(b64Salt, b64Hash string) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argon2Params.memory, argon2Params.iterations, argon2Params.parallelism, b64Salt, b64Hash)
}

func ComparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		slog.Error("compare password and hash:", slog.Any("error", err))
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
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
