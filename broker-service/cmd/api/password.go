package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	// Cost for bcrypt, higher is more secure but slower
	// Recommended minimum is 12
	bcryptCost = 12
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		return err
	}
	return nil
}

// GenerateToken creates a cryptographically secure random token
// Returns the token as a base64 URL-safe string
func GenerateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateSessionToken creates a token suitable for session management
// Uses 32 bytes of entropy (resulting in a 43-character base64 string)
func GenerateSessionToken() (string, error) {
	return GenerateToken(32)
}

// GenerateCSRFToken creates a token suitable for CSRF protection
// Uses 32 bytes of entropy (resulting in a 43-character base64 string)
func GenerateCSRFToken() (string, error) {
	return GenerateToken(32)
}
