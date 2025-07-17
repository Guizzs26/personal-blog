package hashx

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Generate(password string) (string, error)
	Compare(hashedPassword, password string) bool
}

// Generate creates a bcrypt hash from a plain-text password
func Generate(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashx: error generating password hash: %v", err)
	}
	return string(hashedPassword), nil
}

// Compare checks whether a given password matches a bcrypt hash
func Compare(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
