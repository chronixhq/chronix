package agentregister

import (
	cfgpkg "chronix-agent/agentconfig"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os/user"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func BuildJWT(cfg *cfgpkg.Config, version string) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(cfg.PrivKeyB64)
	if err != nil {
		return "", err
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("invalid private key size")
	}
	priv := ed25519.PrivateKey(privBytes)

	meta := CollectMetadata()
	claims := jwt.MapClaims{
		"uuid":      cfg.UUID,
		"name":      cfg.Name,
		"version":   version,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"os":        meta["os"],
		"arch":      meta["arch"],
		"osVersion": meta["os_version"],
		"osType":    meta["os_type"],
	}
	if u, err := user.Current(); err == nil {
		claims["runningUser"] = u.Username
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(priv)
}
