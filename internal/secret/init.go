package secret

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
)

// InitMasterKey initializes the secret package by looking for a key in:
// 1. CHRONIX_MASTER_KEY environment variable.
// 2. A .master_key file in the provided data directory.
// If neither exists, a new key is generated and saved to the file.
func InitMasterKey(dataDir string) {
	key := os.Getenv("CHRONIX_MASTER_KEY")
	if key != "" {
		Setup(key)
		slog.Debug("Master key initialized from environment variable")
		return
	}

	keyPath := filepath.Join(dataDir, ".master_key")
	data, err := os.ReadFile(keyPath)
	if err == nil && len(data) > 0 {
		Setup(string(data))
		slog.Debug("Master key initialized from file", "path", keyPath)
		return
	}

	// Generate a new key
	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		slog.Error("Failed to generate random master key", "error", err)
		// Fallback to a very insecure default if we really can't generate one
		// But it's better to panic here as encryption is critical
		panic("failed to generate master key: " + err.Error())
	}
	keyStr := hex.EncodeToString(newKey)
	if err := os.WriteFile(keyPath, []byte(keyStr), 0600); err != nil {
		slog.Error("Failed to save master key to file", "path", keyPath, "error", err)
	} else {
		slog.Info("Generated new master key and saved to file", "path", keyPath)
	}

	Setup(keyStr)
}
