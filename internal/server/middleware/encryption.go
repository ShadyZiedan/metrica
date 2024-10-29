package middleware

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Encryption struct {
	privateKey *rsa.PrivateKey
}

func NewEncryption(privateKey *rsa.PrivateKey) *Encryption {
	return &Encryption{privateKey: privateKey}
}

func NewEncryptionFromFile(privateKeyPath string) (*Encryption, error) {
	f, err := os.Open(privateKeyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	keyData, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %w", err)
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	return NewEncryption(privateKey), nil
}

func (e *Encryption) MiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		aesKey := r.Header.Get(`X-Encrypted-Key`)
		if aesKey == "" {
			http.Error(w, "missing encryption key", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		decryptedBody, err := e.decryptMessage(aesKey, body)
		if err != nil {
			http.Error(w, "failed to decrypt message", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(decryptedBody))

		next.ServeHTTP(w, r)
	})
}

func (e *Encryption) decryptMessage(encryptedAESKey string, body []byte) ([]byte, error) {
	decodeString, err := base64.StdEncoding.DecodeString(encryptedAESKey)
	if err != nil {
		return nil, err
	}
	decryptedAESKey, err := rsa.DecryptPKCS1v15(rand.Reader, e.privateKey, decodeString)
	if err != nil {
		return nil, err
	}

	body, err = base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, err
	}

	return decryptWithAES(decryptedAESKey, body)
}

// decryptWithAES decrypts data using AES and returns the plaintext.
func decryptWithAES(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
