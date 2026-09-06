// Package crypto provides at-rest encryption for secrets that must be
// stored in plaintext-equivalent form to be usable later (OAuth refresh
// tokens), unlike passwords which are hashed and never need to be reversed.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// deriveKey turns an arbitrary-length configured secret into a 32-byte
// AES-256 key, so TOKEN_ENCRYPTION_KEY can be any non-empty string rather
// than requiring an operator to generate and paste an exact-length key.
func deriveKey(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// Encrypt returns a base64-encoded AES-256-GCM ciphertext (nonce prepended)
// for plaintext, keyed by secret.
func Encrypt(plaintext, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("crypto: encryption secret is empty")
	}
	block, err := aes.NewCipher(deriveKey(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt.
func Decrypt(encoded, secret string) (string, error) {
	if secret == "" {
		return "", errors.New("crypto: encryption secret is empty")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(deriveKey(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
