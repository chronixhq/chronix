package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	masterKey []byte
	Prefix    = "enc:"
)

// Setup initializes the secret package with a master key.
// The key is hashed using SHA-256 to ensure it is 32 bytes.
func Setup(key string) {
	h := sha256.New()
	h.Write([]byte(key))
	masterKey = h.Sum(nil)
}

// Encrypt encrypts plainText using AES-GCM and returns a base64 encoded string with a prefix.
func Encrypt(plainText string) (string, error) {
	if len(masterKey) == 0 {
		return "", errors.New("secret package not initialized")
	}
	if plainText == "" {
		return "", nil
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return Prefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext (base64 encoded with prefix).
// If the string does not have the prefix, it is returned as-is (graceful degradation for plain text).
func Decrypt(ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, Prefix) {
		return ciphertext, nil
	}

	if len(masterKey) == 0 {
		return "", errors.New("secret package not initialized")
	}

	raw := strings.TrimPrefix(ciphertext, Prefix)
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// EncryptPtr encrypts a string pointer.
func EncryptPtr(p *string) (*string, error) {
	if p == nil || *p == "" || strings.HasPrefix(*p, Prefix) {
		return p, nil
	}
	enc, err := Encrypt(*p)
	if err != nil {
		return nil, err
	}
	return &enc, nil
}

// DecryptPtr decrypts a string pointer.
func DecryptPtr(p *string) (*string, error) {
	if p == nil || *p == "" || !strings.HasPrefix(*p, Prefix) {
		return p, nil
	}
	dec, err := Decrypt(*p)
	if err != nil {
		return nil, err
	}
	return &dec, nil
}
