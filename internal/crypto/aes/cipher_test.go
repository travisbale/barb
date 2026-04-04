package aes

import (
	"bytes"
	"testing"
)

func testKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

func TestRoundTrip(t *testing.T) {
	c := NewCipher(testKey())

	encrypted, err := c.Encrypt([]byte("hunter2"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(encrypted, []byte("hunter2")) {
		t.Fatal("encrypted value should not equal plaintext")
	}

	decrypted, err := c.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(decrypted) != "hunter2" {
		t.Errorf("Decrypt = %q, want %q", decrypted, "hunter2")
	}
}

func TestEmpty(t *testing.T) {
	c := NewCipher(testKey())

	encrypted, err := c.Encrypt(nil)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if encrypted != nil {
		t.Errorf("Encrypt(nil) = %v, want nil", encrypted)
	}

	decrypted, err := c.Decrypt(nil)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if decrypted != nil {
		t.Errorf("Decrypt(nil) = %v, want nil", decrypted)
	}
}

func TestUniqueNonces(t *testing.T) {
	c := NewCipher(testKey())

	first, _ := c.Encrypt([]byte("same"))
	second, _ := c.Encrypt([]byte("same"))

	if bytes.Equal(first, second) {
		t.Error("two encryptions of the same plaintext should produce different ciphertext")
	}
}

func TestWrongKeyFails(t *testing.T) {
	encryptor := NewCipher(testKey())
	encrypted, _ := encryptor.Encrypt([]byte("secret"))

	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = byte(i + 1)
	}
	decryptor := NewCipher(wrongKey)

	_, err := decryptor.Decrypt(encrypted)
	if err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestTruncatedCiphertextFails(t *testing.T) {
	c := NewCipher(testKey())

	_, err := c.Decrypt([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for truncated ciphertext")
	}
}
