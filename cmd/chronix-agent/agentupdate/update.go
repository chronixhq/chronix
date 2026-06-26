package agentupdate

import (
	cfgpkg "chronix-agent/agentconfig"
	regpkg "chronix-agent/agentregister"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func targetExePath(currentPath, currentVersion, newVersion string) string {
	dir := filepath.Dir(currentPath)
	base := filepath.Base(currentPath)

	if strings.HasPrefix(base, "agent") {
		base = "chronix-" + base
	}

	if currentVersion != "" && currentVersion != "dev" && strings.Contains(base, currentVersion) {
		newBase := strings.Replace(base, currentVersion, newVersion, 1)
		return filepath.Join(dir, newBase)
	}

	return filepath.Join(dir, base)
}

func ApplyUpdate(cfg *cfgpkg.Config, version string, sha256Expected string, signatureExpected string, currentVersion string, publicKey string, restart func(string) error) error {
	if version == currentVersion {
		slog.Debug("Agent already at target version, skipping update", slog.String("version", version))
		return nil
	}

	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://%s:%d/agent/update/%s/%s", cfg.ServerHost, cfg.WSPort, version, platform)

	jwtStr, err := regpkg.BuildJWT(cfg, currentVersion)
	if err != nil {
		return fmt.Errorf("generate jwt: %w", err)
	}

	slog.Info("Downloading update from relay", slog.String("version", version), slog.String("url", url))

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	nextExePath := targetExePath(exePath, currentVersion, version)
	tempPath := nextExePath + ".new"
	if err := downloadFile(url, jwtStr, tempPath); err != nil {
		return fmt.Errorf("download binary: %w", err)
	}
	defer func() { _ = os.Remove(tempPath) }()

	if sha256Expected != "" {
		if err := verifySHA256(tempPath, sha256Expected); err != nil {
			return fmt.Errorf("verify SHA256: %w", err)
		}
	}

	if signatureExpected != "" && publicKey != "" {
		if err := verifySignature(tempPath, signatureExpected, publicKey); err != nil {
			return fmt.Errorf("verify signature: %w", err)
		}
	}

	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("backup current binary: %w", err)
	}

	if err := os.Rename(tempPath, nextExePath); err != nil {
		_ = os.Rename(oldPath, exePath)
		return fmt.Errorf("install new binary: %w", err)
	}

	if err := os.Chmod(nextExePath, 0755); err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	slog.Info("Update applied successfully. Restarting agent...")
	if err := restart(nextExePath); err != nil {
		return err
	}

	return nil
}

func downloadFile(url string, jwt string, destPath string) error {
	client := &http.Client{
		Timeout: 5 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := client.Do(req)
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
