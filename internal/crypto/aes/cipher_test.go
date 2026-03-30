package aes

import (
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
	cipher := NewCipher(testKey())

	encrypted, err := cipher.Encrypt("hunter2")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if encrypted == "hunter2" {
		t.Fatal("encrypted value should not equal plaintext")
	}

	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if decrypted != "hunter2" {
		t.Errorf("Decrypt = %q, want %q", decrypted, "hunter2")
	}
}

func TestEmptyString(t *testing.T) {
	cipher := NewCipher(testKey())

	encrypted, err := cipher.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if encrypted != "" {
		t.Errorf("Encrypt('') = %q, want empty", encrypted)
	}

	decrypted, err := cipher.Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if decrypted != "" {
		t.Errorf("Decrypt('') = %q, want empty", decrypted)
	}
}

func TestUniqueNonces(t *testing.T) {
	cipher := NewCipher(testKey())

	first, _ := cipher.Encrypt("same")
	second, _ := cipher.Encrypt("same")

	if first == second {
		t.Error("two encryptions of the same plaintext should produce different ciphertext")
	}
}

func TestWrongKeyFails(t *testing.T) {
	encryptor := NewCipher(testKey())
	encrypted, _ := encryptor.Encrypt("secret")

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

func TestInvalidBase64Fails(t *testing.T) {
	cipher := NewCipher(testKey())

	_, err := cipher.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestTruncatedCiphertextFails(t *testing.T) {
	cipher := NewCipher(testKey())

	_, err := cipher.Decrypt("AQID") // valid base64, but too short
	if err == nil {
		t.Error("expected error for truncated ciphertext")
	}
}
