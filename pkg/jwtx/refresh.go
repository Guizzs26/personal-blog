package jwtx

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateRefreshToken creates a cryptographically secure random token.
// It returns both the raw token (to be sent to the client) and a hashed version (to be stored in the database)
func GenerateRefreshToken() (string, string, error) {
	// Generate 64 random bytes
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		return "", "", fmt.Errorf("generate-refresh-token: %v", err)
	}

	// Encode the random bytes to a hex string (safe for transport)
	token := hex.EncodeToString(b)

	// Hash the token using SHA256 before storing in the database
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	return token, hashedToken, nil
}

// HashRefreshToken hashes a given refresh token using SHA256.
// Use this when validating an incoming refresh token against the stored hash
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
