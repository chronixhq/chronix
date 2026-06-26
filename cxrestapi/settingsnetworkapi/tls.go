package settingsnetworkapi

import (
	"bytes"
	cxsettingspkg "chronix/internal/cxsettings"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

func uploadHTTPSCert(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Expected multipart form file 'file'")
		return
	}
	f, err := file.Open()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read uploaded file", err.Error())
		return
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(f)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read uploaded file", err.Error())
		return
	}
	text := string(b)
	if !strings.Contains(text, "BEGIN CERTIFICATE") {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Uploaded file does not appear to be a PEM certificate")
		return
	}
	if len(b) > 100*1024 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Certificate file too large")
		return
	}

	cxsettingspkg.CxSettings.HTTPSCertPem = utilities.Ptr(text)
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving certificate", err.Error())
		return
	}
	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)
	restresponse.RestSuccess(c, gin.H{"certFileName": file.Filename})
}

func uploadHTTPSKey(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Expected multipart form file 'file'")
		return
	}
	f, err := file.Open()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read uploaded file", err.Error())
		return
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(f)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read uploaded file", err.Error())
		return
	}
	text := string(b)
	if !strings.Contains(text, "PRIVATE KEY") {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Uploaded file does not appear to be a PEM private key")
		return
	}
	if len(b) > 100*1024 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Key file too large")
		return
	}

	cxsettingspkg.CxSettings.HTTPSKeyPem = utilities.Ptr(text)
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving key", err.Error())
		return
	}
	restresponse.RestSuccess(c, gin.H{"keyFileName": file.Filename})
}

func uploadHTTPSCertAndKey(c *gin.Context) {
	certFile, certErr := c.FormFile("cert")
	keyFile, keyErr := c.FormFile("key")
	if certErr != nil || keyErr != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Expected multipart form fields 'cert' and 'key'")
		return
	}

	cf, err := certFile.Open()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read certificate", err.Error())
		return
	}
	defer func() { _ = cf.Close() }()
	certBytes, err := io.ReadAll(cf)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read certificate", err.Error())
		return
	}
	if len(certBytes) > 100*1024 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Certificate file too large")
		return
	}

	kf, err := keyFile.Open()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read key", err.Error())
		return
	}
	defer func() { _ = kf.Close() }()
	keyBytes, err := io.ReadAll(kf)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Unable to read key", err.Error())
		return
	}
	if len(keyBytes) > 100*1024 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Key file too large")
		return
	}

	certPEM := string(certBytes)
	keyPEM := string(keyBytes)
	if !strings.Contains(certPEM, "BEGIN CERTIFICATE") {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Uploaded cert does not appear to be a PEM certificate")
		return
	}
	if !strings.Contains(keyPEM, "PRIVATE KEY") {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Uploaded key does not appear to be a PEM private key")
		return
	}
	info, err := validateCertKeyPair(certPEM, keyPEM)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Certificate/key validation failed", err.Error())
		return
	}

	cxsettingspkg.CxSettings.HTTPSCertPem = utilities.Ptr(certPEM)
	cxsettingspkg.CxSettings.HTTPSKeyPem = utilities.Ptr(keyPEM)
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving certificate/key", err.Error())
		return
	}
	restresponse.RestSuccess(c, gin.H{
		"certFileName": certFile.Filename,
		"keyFileName":  keyFile.Filename,
		"certInfo": gin.H{
			"subject":  info.Subject,
			"issuer":   info.Issuer,
			"validity": fmt.Sprintf("%s - %s", info.NotBefore, info.NotAfter),
		},
	})
}

func deleteHTTPSCert(c *gin.Context) {
	cxsettingspkg.CxSettings.HTTPSCertPem = nil
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error removing certificate", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func deleteHTTPSKey(c *gin.Context) {
	cxsettingspkg.CxSettings.HTTPSKeyPem = nil
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error removing key", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func deleteHTTPSCertAndKey(c *gin.Context) {
	cxsettingspkg.CxSettings.HTTPSCertPem = nil
	cxsettingspkg.CxSettings.HTTPSKeyPem = nil
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error removing certificate/key", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

type CertSummary struct {
	Subject   string
	Issuer    string
	NotBefore string
	NotAfter  string
}

func SummarizeCert(pemText string) (*CertSummary, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil || block.Type != "CERTIFICATE" {
		data := []byte(pemText)
		for {
			b, rest := pem.Decode(data)
			if b == nil {
				break
			}
			if b.Type == "CERTIFICATE" {
				block = b
				break
			}
			data = rest
		}
		if block == nil {
			return nil, fmt.Errorf("no CERTIFICATE block found")
		}
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &CertSummary{
		Subject:   cert.Subject.String(),
		Issuer:    cert.Issuer.String(),
		NotBefore: cert.NotBefore.UTC().Format("2 Jan 2006, 15:04 MST"),
		NotAfter:  cert.NotAfter.UTC().Format("2 Jan 2006, 15:04 MST"),
	}, nil
}

func validateCertKeyPair(certPEM, keyPEM string) (*CertSummary, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("invalid certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse cert: %w", err)
	}

	var pubFromKey any
	kb, _ := pem.Decode([]byte(keyPEM))
	if kb == nil {
		return nil, fmt.Errorf("invalid key PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(kb.Bytes); err == nil {
		pubFromKey = &key.PublicKey
	} else if kAny, err := x509.ParsePKCS8PrivateKey(kb.Bytes); err == nil {
		switch k := kAny.(type) {
		case *rsa.PrivateKey:
			pubFromKey = &k.PublicKey
		case *ecdsa.PrivateKey:
			pubFromKey = &k.PublicKey
		}
	} else if key, err := x509.ParseECPrivateKey(kb.Bytes); err == nil {
		pubFromKey = &key.PublicKey
	}
	if pubFromKey == nil {
		summary, _ := SummarizeCert(certPEM)
		return summary, nil
	}

	certPubDER, _ := x509.MarshalPKIXPublicKey(cert.PublicKey)
	keyPubDER, _ := x509.MarshalPKIXPublicKey(pubFromKey)
	if certPubDER != nil && keyPubDER != nil && !bytes.Equal(certPubDER, keyPubDER) {
		return nil, fmt.Errorf("certificate does not match private key")
	}

	summary, _ := SummarizeCert(certPEM)
	return summary, nil
}
