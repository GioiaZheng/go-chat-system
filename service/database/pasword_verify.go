package database

import "crypto/subtle"

// VerifyPassword does a constant-time comparison between input and stored.
// Replace with bcrypt/argon2 in production.
func VerifyPassword(input, stored string) bool {
	return subtle.ConstantTimeCompare([]byte(input), []byte(stored)) == 1
}
