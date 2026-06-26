package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

func getBinaryMetadata(binaryPath, privHex string) (BinaryInfo, error) {
	file, err := os.Open(binaryPath)
	if err != nil {
		return BinaryInfo{}, err
	}
	defer func() { _ = file.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return BinaryInfo{}, err
	}
	sha256Hex := hex.EncodeToString(h.Sum(nil))

	privBytes, _ := hex.DecodeString(privHex)
	if _, err := file.Seek(0, 0); err != nil {
		return BinaryInfo{}, err
	}
	content, err := io.ReadAll(file)
	if err != nil {
		return BinaryInfo{}, err
	}
	sig := ed25519.Sign(privBytes, content)
	sigHex := hex.EncodeToString(sig)

	binaryName := filepath.Base(binaryPath)
	binaryURL := "https://dist.chronixhq.com/downloads/" + binaryName

	return BinaryInfo{
		URL:       binaryURL,
		SHA256:    sha256Hex,
		Signature: sigHex,
	}, nil
}
