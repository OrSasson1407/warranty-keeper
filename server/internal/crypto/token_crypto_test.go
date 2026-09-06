package crypto_test

import (
	"testing"

	"warrantykeeper/server/internal/crypto"
)

func TestEncryptDecrypt_RoundTrips(t *testing.T) {
	ciphertext, err := crypto.Encrypt("a-refresh-token", "secret-key")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if ciphertext == "a-refresh-token" {
		t.Error("ciphertext should not equal the plaintext")
	}

	plaintext, err := crypto.Decrypt(ciphertext, "secret-key")
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}
	if plaintext != "a-refresh-token" {
		t.Errorf("Decrypt = %q, want %q", plaintext, "a-refresh-token")
	}
}

func TestEncrypt_DifferentCiphertextEachTime(t *testing.T) {
	a, err := crypto.Encrypt("same-plaintext", "secret-key")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	b, err := crypto.Encrypt("same-plaintext", "secret-key")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if a == b {
		t.Error("expected different ciphertexts for the same plaintext due to random nonce")
	}
}

func TestDecrypt_WrongKeyFails(t *testing.T) {
	ciphertext, err := crypto.Encrypt("a-refresh-token", "secret-key")
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if _, err := crypto.Decrypt(ciphertext, "wrong-key"); err == nil {
		t.Error("expected an error decrypting with the wrong key")
	}
}

func TestEncrypt_EmptySecretFails(t *testing.T) {
	if _, err := crypto.Encrypt("x", ""); err == nil {
		t.Error("expected an error encrypting with an empty secret")
	}
}

func TestDecrypt_EmptySecretFails(t *testing.T) {
	if _, err := crypto.Decrypt("x", ""); err == nil {
		t.Error("expected an error decrypting with an empty secret")
	}
}
