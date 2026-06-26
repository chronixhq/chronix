// Package main provides a tool for signing and verifying Chronix binaries.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run dev/sign.go gen-keys [prefix]")
		fmt.Println("  go run dev/sign.go sign <private_key_hex> <file_path>")
		fmt.Println("  go run dev/sign.go sign-file <private_key_file> <file_path>")
		fmt.Println("  go run dev/sign.go verify <public_key_hex> <signature_hex> <file_path>")
		fmt.Println("  go run dev/sign.go verify-file <public_key_file> <signature_hex> <file_path>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "gen-keys":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(os.Args) > 2 {
			prefix := os.Args[2]
			privFile := prefix + ".key"
			pubFile := prefix + ".pub"

			err = os.WriteFile(privFile, []byte(hex.EncodeToString(priv)), 0600)
			if err != nil {
				fmt.Printf("Error saving private key: %v\n", err)
				os.Exit(1)
			}
			err = os.WriteFile(pubFile, []byte(hex.EncodeToString(pub)), 0644)
			if err != nil {
				fmt.Printf("Error saving public key: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Keys saved to %s and %s\n", privFile, pubFile)
			fmt.Printf("Public Key (Hex):  %x\n", pub)
		} else {
			fmt.Printf("Public Key (Hex):  %x\n", pub)
			fmt.Printf("Private Key (Hex): %x\n", priv)
		}

	case "sign":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run dev/sign.go sign <private_key_hex> <file_path>")
			os.Exit(1)
		}
		sign(os.Args[2], os.Args[3])

	case "sign-file":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run dev/sign.go sign-file <private_key_file> <file_path>")
			os.Exit(1)
		}
		keyData, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error reading key file: %v\n", err)
			os.Exit(1)
		}
		sign(string(keyData), os.Args[3])

	case "verify":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run dev/sign.go verify <public_key_hex> <signature_hex> <file_path>")
			os.Exit(1)
		}
		verify(os.Args[2], os.Args[3], os.Args[4])

	case "verify-file":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run dev/sign.go verify-file <public_key_file> <signature_hex> <file_path>")
			os.Exit(1)
		}
		keyData, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Printf("Error reading key file: %v\n", err)
			os.Exit(1)
		}
		verify(string(keyData), os.Args[3], os.Args[4])

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func sign(privHex, filePath string) {
	privBytes, err := hex.DecodeString(privHex)
	if err != nil {
		fmt.Printf("Error decoding private key: %v\n", err)
		os.Exit(1)
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		fmt.Printf("Error: private key must be %d bytes (got %d)\n", ed25519.PrivateKeySize, len(privBytes))
		os.Exit(1)
	}

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()

	content, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	sig := ed25519.Sign(privBytes, content)
	fmt.Printf("Signature (Hex): %x\n", sig)
}

func verify(pubHex, sigHex, filePath string) {
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		fmt.Printf("Error decoding public key: %v\n", err)
		os.Exit(1)
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		fmt.Printf("Error decoding signature: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()

	content, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	if ed25519.Verify(pubBytes, content, sigBytes) {
		fmt.Println("Signature is VALID")
	} else {
		fmt.Println("Signature is INVALID")
		os.Exit(1)
	}
}
