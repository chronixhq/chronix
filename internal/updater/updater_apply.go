package updater

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func ApplyUpdate(manifest *VersionManifest, shouldRestart bool, currentVersion string) error {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	binaryInfo, ok := manifest.Server.Binaries[platform]
	if !ok {
		return fmt.Errorf("no binary available for platform %s", platform)
	}

	slog.Info("Applying update", slog.String("version", manifest.Server.Version), slog.String("url", binaryInfo.URL))

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	targetExePath := getTargetExePath(exePath, currentVersion, manifest.Server.Version)
	tempPath := targetExePath + ".new"
	if err := downloadFile(binaryInfo.URL, tempPath); err != nil {
		return fmt.Errorf("download binary: %w", err)
	}
	defer func() { _ = os.Remove(tempPath) }()

	if err := verifySHA256(tempPath, binaryInfo.SHA256); err != nil {
		return fmt.Errorf("verify SHA256: %w", err)
	}
	if len(PublicKey) > 0 {
		if err := verifySignature(tempPath, binaryInfo.Signature, PublicKey); err != nil {
			return fmt.Errorf("verify signature: %w", err)
		}
	}

	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("backup current binary: %w", err)
	}

	if err := os.Rename(tempPath, targetExePath); err != nil {
		_ = os.Rename(oldPath, exePath)
		return fmt.Errorf("install new binary: %w", err)
	}
	if err := os.Chmod(targetExePath, 0755); err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	slog.Info("Update applied successfully.")
	if shouldRestart {
		if RecordActivityHook != nil {
			RecordActivityHook("Server Update Applied", "Updated to version "+manifest.Server.Version)
		}
		if err := Restart(targetExePath); err != nil {
			slog.Error("Failed to restart after update", "error", err)
			os.Exit(1)
		}
	}
	return nil
}

func GetAgentBinaryPath(version, platform string) string {
	return filepath.Join(DataDir, "updates", "agent", version, "chronix-agent-"+platform)
}

func EnsureAgentBinaryCached(manifest *VersionManifest, platform string) (string, error) {
	binaryInfo, ok := manifest.Agent.Binaries[platform]
	if !ok {
		return "", fmt.Errorf("no agent binary available for platform %s", platform)
	}

	destPath := GetAgentBinaryPath(manifest.Agent.Version, platform)
	if _, err := os.Stat(destPath); err == nil {
		if err := verifySHA256(destPath, binaryInfo.SHA256); err == nil {
			return destPath, nil
		}
		slog.Warn("cached agent binary hash mismatch, re-downloading", slog.String("path", destPath))
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return "", fmt.Errorf("create cache directory: %w", err)
	}

	slog.Info("Caching agent binary", slog.String("version", manifest.Agent.Version), slog.String("platform", platform), slog.String("url", binaryInfo.URL))
	tempPath := destPath + ".tmp"
	if err := downloadFile(binaryInfo.URL, tempPath); err != nil {
		return "", fmt.Errorf("download agent binary: %w", err)
	}
	defer func() { _ = os.Remove(tempPath) }()

	if err := verifySHA256(tempPath, binaryInfo.SHA256); err != nil {
		return "", fmt.Errorf("verify agent binary SHA256: %w", err)
	}
	if len(PublicKey) > 0 {
		if err := verifySignature(tempPath, binaryInfo.Signature, PublicKey); err != nil {
			return "", fmt.Errorf("verify agent binary signature: %w", err)
		}
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		return "", fmt.Errorf("save cached agent binary: %w", err)
	}
	return destPath, nil
}

func downloadFile(url string, destPath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifySHA256(filePath string, expectedHash string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(h.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}
	return nil
}

func verifySignature(filePath string, signatureHex string, publicKeyHex string) error {
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return errors.New("invalid public key size")
	}

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}
	if len(sigBytes) != ed25519.SignatureSize {
		return errors.New("invalid signature size")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	if !ed25519.Verify(pubKeyBytes, content, sigBytes) {
		return errors.New("signature verification failed")
	}
	return nil
}
