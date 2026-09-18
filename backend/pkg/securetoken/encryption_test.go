package securetoken

import (
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte(strings.Repeat("a", 32))

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	original := "example-refresh-token"

	encrypted, err := encryptor.Encrypt(original)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	if encrypted == original {
		t.Fatalf("encrypted value must differ from plaintext")
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if decrypted != original {
		t.Fatalf("decrypted value mismatch: got %q, want %q", decrypted, original)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := []byte(strings.Repeat("a", 32))
	key2 := []byte(strings.Repeat("b", 32))

	encryptor1, err := NewEncryptor(key1)
	if err != nil {
		t.Fatalf("failed to create first encryptor: %v", err)
	}

	encryptor2, err := NewEncryptor(key2)
	if err != nil {
		t.Fatalf("failed to create second encryptor: %v", err)
	}

	encrypted, err := encryptor1.Encrypt("secret-token")
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	if _, err := encryptor2.Decrypt(encrypted); err == nil {
		t.Fatal("expected decryption with wrong key to fail")
	}

}

func TestInvalidKeyLength(t *testing.T) {
	_, err := NewEncryptor([]byte("short-key"))

	if err == nil {
		t.Fatal("expected invalid key length error")
	}
}
