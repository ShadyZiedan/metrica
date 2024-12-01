package middleware

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper function to generate a mock RSA private key
func generatePrivateKey(t *testing.T) *rsa.PrivateKey {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	return priv
}

// Helper function to create a test server with the middleware
func createTestServer(middleware http.Handler) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("/", middleware)
	return httptest.NewServer(mux)
}

func TestMiddleware_MissingEncryptionKey(t *testing.T) {
	// Setup a private key for decryption
	privateKey := generatePrivateKey(t)
	enc := NewEncryption(privateKey)

	// Create a test handler that we will apply the middleware to
	handler := enc.MiddleWare(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test server
	ts := createTestServer(handler)
	defer ts.Close()

	// Make a request without the X-Encrypted-Key header
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader([]byte(`{"data": "test"}`)))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMiddleware_SuccessfulDecryption(t *testing.T) {
	// Setup a private key for decryption
	privateKey := generatePrivateKey(t)
	enc := NewEncryption(privateKey)

	// Create an AES key and encrypt some data (for simplicity, we mock it here)
	aesKey := []byte("thisisaverysecretkey32bytes12345") // 32-byte AES key (for AES-256)
	originalData := []byte(`{"data": "test"}`)

	// Encrypt the data with AES
	encryptedData, err := encryptWithAES(aesKey, originalData)
	assert.NoError(t, err)

	// Encrypt the AES key with RSA
	encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, aesKey)
	assert.NoError(t, err)
	encodedEncryptedAESKey := base64.StdEncoding.EncodeToString(encryptedAESKey)
	encodedEncryptedData := base64.StdEncoding.EncodeToString(encryptedData)

	// Create a test handler that will receive the decrypted data
	handler := enc.MiddleWare(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, originalData, body)
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test server
	ts := createTestServer(handler)
	defer ts.Close()

	// Make a request with the encrypted AES key in the header and encrypted data in the body
	req, err := http.NewRequest("POST", ts.URL, bytes.NewReader([]byte(encodedEncryptedData)))
	assert.NoError(t, err)
	req.Header.Set("X-Encrypted-Key", encodedEncryptedAESKey)

	resp, err := ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMiddleware_DecryptionFailure(t *testing.T) {
	// Setup a private key for decryption
	privateKey := generatePrivateKey(t)
	enc := NewEncryption(privateKey)

	// Create an AES key and encrypt some data
	aesKey := []byte("thisisaverysecretkey32bytes12345") // 32-byte AES key
	originalData := []byte(`{"data": "test"}`)

	// Encrypt the data with AES
	encryptedData, err := encryptWithAES(aesKey, originalData)
	assert.NoError(t, err)

	// Encrypt the AES key with RSA
	encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, aesKey)
	assert.NoError(t, err)
	encodedEncryptedAESKey := base64.StdEncoding.EncodeToString(encryptedAESKey)
	encodedEncryptedData := base64.StdEncoding.EncodeToString(encryptedData)

	// Create a test handler
	handler := enc.MiddleWare(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test server
	ts := createTestServer(handler)
	defer ts.Close()

	// Modify the encrypted data to simulate a decryption failure
	invalidEncryptedData := encodedEncryptedData + "corrupt"

	// Make a request with the modified encrypted data
	req, err := http.NewRequest("POST", ts.URL, bytes.NewReader([]byte(invalidEncryptedData)))
	assert.NoError(t, err)
	req.Header.Set("X-Encrypted-Key", encodedEncryptedAESKey)

	resp, err := ts.Client().Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func encryptWithAES(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}
