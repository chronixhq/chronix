package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadOrGenKeys(path string, pathSupplied bool) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read key file: %w", err)
		}

		var kp KeyPair
		if err := json.Unmarshal(data, &kp); err != nil {
			return nil, nil, fmt.Errorf("invalid key file format: %w", err)
		}

		privBytes, err := hex.DecodeString(kp.Private)
		if err != nil {
			return nil, nil, fmt.Errorf("decode private key: %w", err)
		}

		pub, err := hex.DecodeString(kp.Public)
		if err != nil {
			return nil, nil, fmt.Errorf("decode public key: %w", err)
		}

		if len(privBytes) != ed25519.PrivateKeySize {
			return nil, nil, fmt.Errorf("invalid private key size: %d", len(privBytes))
		}

		priv := ed25519.PrivateKey(privBytes)
		expectedPub := priv.Public().(ed25519.PublicKey)
		if hex.EncodeToString(expectedPub) != kp.Public {
			return nil, nil, fmt.Errorf("public key does not match private key")
		}

		return priv, pub, nil
	}

	if pathSupplied {
		return nil, nil, fmt.Errorf("key file not found: %s", path)
	}

	fmt.Printf("Generating new key pair at %s...\n", path)
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate keys: %w", err)
	}

	kp := KeyPair{
		Private: hex.EncodeToString(priv),
		Public:  hex.EncodeToString(pub),
	}

	data, _ := json.MarshalIndent(kp, "", "  ")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, nil, fmt.Errorf("write key file: %w", err)
	}

	return priv, pub, nil
}
