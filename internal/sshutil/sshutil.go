package sshutil

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"golang.org/x/crypto/ssh"
)

type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

func GenerateED25519KeyPair(format string) (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	var privStr string
	if strings.ToLower(format) == "pkcs8" || strings.ToLower(format) == "pem" {
		// PKCS#8 format (traditional PEM)
		pkcs8Priv, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			return nil, err
		}
		privStr = string(pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pkcs8Priv,
		}))
	} else {
		// OpenSSH format
		privBlock, err := ssh.MarshalPrivateKey(priv, "")
		if err != nil {
			return nil, err
		}
		privStr = string(pem.EncodeToMemory(privBlock))
	}

	// Public key in OpenSSH format (usually what is needed for authorized_keys)
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, err
	}
	pubStr := string(ssh.MarshalAuthorizedKey(sshPub))

	return &KeyPair{
		PrivateKey: privStr,
		PublicKey:  pubStr,
	}, nil
}
