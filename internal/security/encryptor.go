package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

type DefaultEncryptor struct {
	pubKey *rsa.PublicKey
	aesKey []byte
}

func NewDefaultEncryptor(pubKey *rsa.PublicKey) (*DefaultEncryptor, error) {
	aesKey := make([]byte, 32)
	_, err := rand.Read(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random AES key: %w", err)
	}
	return &DefaultEncryptor{pubKey: pubKey, aesKey: aesKey}, nil
}

func NewDefaultEncryptorFromFile(path string) (*DefaultEncryptor, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", path, err)
	}
	defer f.Close()

	keyData, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block containing the public key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	encryptor, err := NewDefaultEncryptor(rsaPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}
	return encryptor, nil
}

func (e *DefaultEncryptor) Encrypt(data []byte) ([]byte, error) {
	// Create a new AES cipher block
	block, err := aes.NewCipher(e.aesKey)
	if err != nil {
		return nil, err
	}

	// Use GCM (Galois/Counter Mode) for AES, which provides encryption and authentication
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce for AES-GCM. The nonce should be unique for each encryption with this key.
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the plaintext and append the nonce at the beginning of the ciphertext.
	// GCM seals and adds an authentication tag automatically.
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (e *DefaultEncryptor) GetEncryptedKey() (string, error) {
	encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, e.pubKey, e.aesKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedAESKey), nil
}
