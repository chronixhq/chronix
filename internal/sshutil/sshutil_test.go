package sshutil

import (
	"strings"
	"testing"
)

func TestGenerateED25519KeyPair(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"OpenSSH", "openssh"},
		{"PEM", "pem"},
		{"PKCS8", "pkcs8"},
		{"Default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kp, err := GenerateED25519KeyPair(tt.format)
			if err != nil {
				t.Fatalf("failed to generate key pair: %v", err)
			}
			if kp.PrivateKey == "" {
				t.Error("private key is empty")
			}
			if kp.PublicKey == "" {
				t.Error("public key is empty")
			}

			if strings.ToLower(tt.format) == "pem" || strings.ToLower(tt.format) == "pkcs8" {
				if !strings.Contains(kp.PrivateKey, "BEGIN PRIVATE KEY") {
					t.Errorf("expected PKCS#8 header, got: %s", kp.PrivateKey)
				}
			} else {
				if !strings.Contains(kp.PrivateKey, "BEGIN OPENSSH PRIVATE KEY") {
					t.Errorf("expected OpenSSH header, got: %s", kp.PrivateKey)
				}
			}

			if !strings.HasPrefix(kp.PublicKey, "ssh-ed25519") {
				t.Errorf("expected ssh-ed25519 prefix, got: %s", kp.PublicKey)
			}
		})
	}
}
