package agentregister

import (
	cfgpkg "chronix-agent/agentconfig"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func baseURL(cfg *cfgpkg.Config) string {
	return fmt.Sprintf("https://%s:%d", cfg.ServerHost, cfg.WSPort)
}

func newHTTPClient(timeout time.Duration, expectedPin string) *http.Client {
	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	expectedPin = strings.TrimSpace(expectedPin)
	if expectedPin != "" {
		tlsCfg.VerifyConnection = func(cs tls.ConnectionState) error {
			pin, err := pinFromTLSState(cs)
			if err != nil {
				return err
			}
			if pin != expectedPin {
				return fmt.Errorf("server pin mismatch")
			}
			return nil
		}
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
	}
}

func rememberServerPin(cfg *cfgpkg.Config, resp *http.Response) {
	if cfg == nil || resp == nil || resp.TLS == nil {
		return
	}
	pin, err := pinFromTLSState(*resp.TLS)
	if err != nil || pin == "" {
		return
	}
	cfg.ServerSPKIB64 = pin
}

func pinFromTLSState(cs tls.ConnectionState) (string, error) {
	if len(cs.PeerCertificates) == 0 {
		return "", fmt.Errorf("no peer certificates")
	}
	spki := cs.PeerCertificates[0].RawSubjectPublicKeyInfo
	hash := sha256.Sum256(spki)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}
