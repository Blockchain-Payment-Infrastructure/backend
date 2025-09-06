package tests

import (
	"errors"
	"strings"
	"testing"

	"backend/internal/utils"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOK   bool
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "Abcdef1!",
			wantOK:   true,
			wantErr:  nil,
		},
		{
			name:     "non-ASCII characters",
			password: "Abcdef!😊",
			wantOK:   false,
			wantErr:  utils.ErrorInvalidCharactersInPassword,
		},
		{
			name:     "too short",
			password: "Ab1!",
			wantOK:   false,
			wantErr:  utils.ErrorShortPassword,
		},
		{
			name:     "password exactly at max length",
			password: "Aa1!" + strings.Repeat("a", 58),
			wantOK:   true,
			wantErr:  nil,
		},
		{
			name:     "password exceeding max length",
			password: "Aa1!" + strings.Repeat("a", 69),
			wantOK:   false,
			wantErr:  utils.ErrorPasswordTooLong,
		},
		{
			name:     "missing special character",
			password: "Abcdefgh",
			wantOK:   false,
			wantErr:  utils.ErrorNotEnoughSpecialCharacters,
		},
		{
			name:     "missing uppercase character",
			password: "abcdefg!",
			wantOK:   false,
			wantErr:  utils.ErrorNotEnoughUpperCaseCharacters,
		},
		{
			name:     "missing lowercase character",
			password: "ABCDEFG!",
			wantOK:   false,
			wantErr:  utils.ErrorNotEnoughLowerCaseCharacters,
		},
		{
			name:     "missing digit",
			password: "Abcdefg!",
			wantOK:   false,
			wantErr:  utils.ErrorNotEnoughDigits,
		}, {
			name:     "contains bell control character",
			password: "Aa1!fiwurgireg" + string(rune(7)),
			wantOK:   false,
			wantErr:  utils.ErrorInvalidCharactersInPassword,
		},
		{
			name:     "contains DEL character",
			password: "Aa1!agataega" + string(rune(127)),
			wantOK:   false,
			wantErr:  utils.ErrorInvalidCharactersInPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOK, gotErr := utils.ValidatePassword(tt.password)
			if gotOK != tt.wantOK {
				t.Errorf("ValidatePassword() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("ValidatePassword() gotErr = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestComparePasswordAndHash(t *testing.T) {
	pWord := "StrongP@ssw0rd"

	// Hash the password
	hash := utils.HashPassword(pWord)

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected hash to start with $argon2id$, got %s", hash)
	}

	// Compare with the correct password
	match, err := utils.ComparePasswordAndHash(pWord, hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Fatalf("expected password to match hash")
	}

	// Compare with wrong password
	match, err = utils.ComparePasswordAndHash("WrongPassword123", hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Fatalf("expected password NOT to match hash")
	}
}

func TestComparePasswordAndHash_InvalidHashFormat(t *testing.T) {
	_, err := utils.ComparePasswordAndHash("pass", "invalidhash")
	if err == nil || !errors.Is(err, utils.ErrInvalidHash) {
		t.Fatalf("expected ErrInvalidHash, got %v", err)
	}
}

func TestComparePasswordAndHash_WrongVersion(t *testing.T) {
	pWord := "StrongP@ssw0rd"
	hash := utils.HashPassword(pWord)

	// Tamper with version number in hash
	badHash := strings.Replace(hash, "v=19", "v=999", 1)

	_, err := utils.ComparePasswordAndHash(pWord, badHash)
	if err == nil || !errors.Is(err, utils.ErrIncompatibleVersion) {
		t.Fatalf("expected ErrIncompatibleVersion, got %v", err)
	}
}

func TestComparePasswordAndHash_CorruptedSalt(t *testing.T) {
	pWord := "StrongP@ssw0rd"
	hash := utils.HashPassword(pWord)

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Fatalf("unexpected hash format")
	}

	// Replace salt with invalid base64
	parts[4] = "!!!invalid_base64$$$"
	badHash := strings.Join(parts, "$")

	_, err := utils.ComparePasswordAndHash(pWord, badHash)
	if err == nil {
		t.Fatalf("expected error for corrupted salt")
	}
}

func TestComparePasswordAndHash_CorruptedHash(t *testing.T) {
	pWord := "StrongP@ssw0rd"
	hash := utils.HashPassword(pWord)

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Fatalf("unexpected hash format")
	}

	// Replace hash with invalid base64
	parts[5] = "###bad_base64###"
	badHash := strings.Join(parts, "$")

	_, err := utils.ComparePasswordAndHash(pWord, badHash)
	if err == nil {
		t.Fatalf("expected error for corrupted hash")
	}
}
