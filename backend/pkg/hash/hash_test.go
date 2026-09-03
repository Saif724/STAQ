package hash

import "testing"

func TestHashPassword(t *testing.T) {
	password := "my-secure-password"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hashed == "" {
		t.Fatal("expected non-empty hash")
	}

	if hashed == password {
		t.Fatal("password must not be stored as plaintext")
	}
}

func TestComparePassword(t *testing.T) {
	password := "my-secure-password"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !ComparePassword(password, hashed) {
		t.Fatal("correct password was rejected")
	}

	if ComparePassword("wrong-password", hashed) {
		t.Fatal("wrong password was accepted")
	}
}

func TestDifferentHashesForSamePassword(t *testing.T) {
	password := "my-secure-password"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to create first hash: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to create second hash: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("same password produced identical hashes")
	}
}
