package secret

import (
	"testing"
)

func TestSecret(t *testing.T) {
	key := "test-master-key"
	Setup(key)

	plain := "hello world"
	enc, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if enc == plain {
		t.Fatal("Encrypted text should be different from plain text")
	}

	dec, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if dec != plain {
		t.Fatalf("Decrypted text %q != plain text %q", dec, plain)
	}

	// Test Decrypt without prefix
	plain2 := "not-encrypted"
	dec2, err := Decrypt(plain2)
	if err != nil {
		t.Fatalf("Decrypt failed for non-prefixed: %v", err)
	}
	if dec2 != plain2 {
		t.Fatalf("Decrypt should return original for non-prefixed")
	}

	// Test Ptr versions
	p := &plain
	encP, err := EncryptPtr(p)
	if err != nil {
		t.Fatalf("EncryptPtr failed: %v", err)
	}
	decP, err := DecryptPtr(encP)
	if err != nil {
		t.Fatalf("DecryptPtr failed: %v", err)
	}
	if *decP != plain {
		t.Fatalf("DecryptPtr failed to recover plain text")
	}
}
