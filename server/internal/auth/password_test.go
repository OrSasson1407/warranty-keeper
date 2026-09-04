package auth_test

import (
	"testing"

	"warrantykeeper/server/internal/auth"
)

func TestHashPassword_ProducesAHashDistinctFromThePlaintext(t *testing.T) {
	hash, err := auth.HashPassword("supersecret1")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected a non-empty hash")
	}
	if hash == "supersecret1" {
		t.Error("the hash must not equal the plaintext password")
	}
}

func TestHashPassword_SamePasswordProducesDifferentHashesEachTime(t *testing.T) {
	// bcrypt salts each hash, so two hashes of the same password should
	// differ even though both validate against it.
	hash1, err := auth.HashPassword("supersecret1")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	hash2, err := auth.HashPassword("supersecret1")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected two hashes of the same password to differ (bcrypt salting)")
	}
	if !auth.CheckPassword(hash1, "supersecret1") || !auth.CheckPassword(hash2, "supersecret1") {
		t.Error("both hashes should still validate against the original password")
	}
}

func TestCheckPassword_CorrectAndIncorrectPasswords(t *testing.T) {
	hash, err := auth.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !auth.CheckPassword(hash, "correct-horse-battery-staple") {
		t.Error("CheckPassword = false for the correct password, want true")
	}
	if auth.CheckPassword(hash, "wrong-password") {
		t.Error("CheckPassword = true for an incorrect password, want false")
	}
	if auth.CheckPassword(hash, "") {
		t.Error("CheckPassword = true for an empty password, want false")
	}
}

func TestCheckPassword_RejectsGarbageHash(t *testing.T) {
	if auth.CheckPassword("not-a-real-bcrypt-hash", "anything") {
		t.Error("CheckPassword = true against a malformed hash, want false")
	}
}
