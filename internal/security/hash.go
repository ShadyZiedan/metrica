// Package security provides utility for security related functions.
package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Hash generates a SHA256 HMAC hash of the given value using the provided key.
// The resulting hash is then encoded as a hexadecimal string.
//
// Parameters:
// - value: The byte slice to be hashed.
// - key: The secret key used for the HMAC operation.
//
// Returns:
// - A string representing the hashed value encoded as a hexadecimal string.
// - An error if there was an issue with the hashing process.
func Hash(value []byte, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(value)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
