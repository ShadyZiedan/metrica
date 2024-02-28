package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(value []byte, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(value)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
