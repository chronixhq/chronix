package cxrestapi

import (
	cxsettingspkg "chronix/internal/cxsettings"
	notifypkg "chronix/internal/notify"
	serverruntime "chronix/internal/serverruntime"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dan-sherwin/go-rest-api-server"
	"github.com/dan-sherwin/go-utilities"
)

type listenerConfig struct {
	httpEnabled  bool
	httpsEnabled bool
	agentEnabled bool
	httpPort     int
	httpsPort    int
	agentPort    int
}

func recoverHTTPStartPanic() {
	if r := recover(); r != nil {
		msg := ""
		switch v := r.(type) {
		case error:
			msg = v.Error()
		case string:
			msg = v
		default:
			msg = "unknown panic"
		}
		data := map[string]any{"message": msg}
		notifypkg.TryCreateNotification(notifypkg.CategorySystem, notifypkg.SeverityError, "HTTP server failed to start", nil, &data)
		panic(r)
	}
}

func registerHTTPRouters() {
	restapi.RegisterRouters(anonymousRoutes, authenticatedRoutes)
	restapi.LogLevel = restapi.LevelDebug
	restapi.LogGetRequests = false
}

func currentListenerConfig() listenerConfig {
	cfg := listenerConfig{
		httpEnabled:  false,
		httpsEnabled: true,
		agentEnabled: true,
		httpPort:     DefaultHTTPPort,
		httpsPort:    DefaultHTTPSPort,
		agentPort:    DefaultAgentPort,
	}

	if cxsettingspkg.CxSettings.HTTPSEnabled != nil {
		cfg.httpsEnabled = *cxsettingspkg.CxSettings.HTTPSEnabled
	}
	if cxsettingspkg.CxSettings.HTTPEnabled != nil {
		cfg.httpEnabled = *cxsettingspkg.CxSettings.HTTPEnabled
	}
	if cxsettingspkg.CxSettings.AgentEnabled != nil {
		cfg.agentEnabled = *cxsettingspkg.CxSettings.AgentEnabled
	}
	if cxsettingspkg.CxSettings.HTTPSPort != nil && *cxsettingspkg.CxSettings.HTTPSPort > 0 {
		cfg.httpsPort = int(*cxsettingspkg.CxSettings.HTTPSPort)
	}
	if cxsettingspkg.CxSettings.HTTPPort != nil && *cxsettingspkg.CxSettings.HTTPPort > 0 {
		cfg.httpPort = int(*cxsettingspkg.CxSettings.HTTPPort)
	}
	if cxsettingspkg.CxSettings.AgentPort != nil && *cxsettingspkg.CxSettings.AgentPort > 0 {
		cfg.agentPort = int(*cxsettingspkg.CxSettings.AgentPort)
	}
	if serverruntime.CurrentServerStatus == serverruntime.StatusUninitialized {
		cfg.httpEnabled = true
		cfg.httpsEnabled = true
	}
	if CLIHTTPConfig.DisableHTTP {
		cfg.httpEnabled = false
	}
	if CLIHTTPConfig.DisableHTTPS {
		cfg.httpsEnabled = false
	}
	if CLIHTTPConfig.ForceHTTPPort > 0 {
		cfg.httpPort = int(CLIHTTPConfig.ForceHTTPPort)
	}
	if CLIHTTPConfig.ForceHTTPSPort > 0 {
		cfg.httpsPort = int(CLIHTTPConfig.ForceHTTPSPort)
	}
	if CLIHTTPConfig.ForceAgentPort > 0 {
		cfg.agentPort = int(CLIHTTPConfig.ForceAgentPort)
	}
	return cfg
}

func applyListenerConfig(cfg listenerConfig) {
	agentListeningAddress = fmt.Sprintf("0.0.0.0:%d", cfg.agentPort)
	restapi.HTTPSListeningAddresses = []string{}
	if cfg.agentEnabled {
		restapi.HTTPSListeningAddresses = append(restapi.HTTPSListeningAddresses, agentListeningAddress)
	}
	if cfg.httpsEnabled {
		restapi.HTTPSListeningAddresses = append(restapi.HTTPSListeningAddresses, fmt.Sprintf("0.0.0.0:%d", cfg.httpsPort))
	}
	if cfg.httpEnabled && cfg.httpPort > 0 {
		restapi.ListeningAddresses = []string{fmt.Sprintf("0.0.0.0:%d", cfg.httpPort)}
	} else {
		restapi.ListeningAddresses = []string{}
	}
}

func startConfiguredListeners(cfg listenerConfig, action string) {
	applyListenerConfig(cfg)

	if certFile, keyFile, ok := findDevCertPair(); ok {
		slog.Info(action+" HTTPS with local development certs", slog.String("cert", certFile), slog.String("key", keyFile))
		if cfg.httpEnabled && cfg.httpPort > 0 {
			slog.Info(action+" HTTP listener", slog.String("addr", firstListeningAddress()))
			restapi.StartHttpServer()
		}
		if cfg.httpsEnabled || cfg.agentEnabled {
			restapi.StartHttpsServer(certFile, keyFile, false)
		}
		return
	}

	if cfg.httpEnabled && cfg.httpPort > 0 {
		slog.Info(action+" HTTP listener", slog.String("addr", firstListeningAddress()))
		restapi.StartHttpServer()
	}
	if cfg.httpsEnabled || cfg.agentEnabled {
		startHTTPSListener()
	}
}

func firstListeningAddress() string {
	if len(restapi.ListeningAddresses) > 0 {
		return restapi.ListeningAddresses[0]
	}
	return ""
}

func findDevCertPair() (string, string, bool) {
	if os.Getenv("DEVMODE") != "true" {
		return "", "", false
	}
	certDir := "/Users/dsherwin/certs"
	certFile := certDir + "/localhost+3.pem"
	keyFile := certDir + "/localhost+3-key.pem"
	if _, err1 := os.Stat(certFile); err1 == nil {
		if _, err2 := os.Stat(keyFile); err2 == nil {
			return certFile, keyFile, true
		}
	}
	slog.Warn("Development mode requested but no cert pair found; falling back to auto self-signed", slog.String("dir", certDir))
	return "", "", false
}

func startHTTPSListener() {
	if cxsettingspkg.CxSettings.HTTPSMode == nil {
		slog.Warn("HTTPS mode not set; falling back to auto self-signed")
		cxsettingspkg.CxSettings.HTTPSMode = utilities.Ptr("selfsigned")
	}
	if cxsettingspkg.CxSettings.HTTPSMode != nil && *cxsettingspkg.CxSettings.HTTPSMode == "selfsigned" {
		certFile := filepath.Join(serverruntime.DataDir, "cert.pem")
		keyFile := filepath.Join(serverruntime.DataDir, "key.pem")
		if err := serverruntime.EnsureSelfSignedCert(certFile, keyFile); err != nil {
			slog.Error("failed to ensure self-signed cert", "error", err)
		}
		restapi.StartHttpsServer(certFile, keyFile, false)
		return
	}
	if cxsettingspkg.CxSettings.HTTPSKeyPem == nil || cxsettingspkg.CxSettings.HTTPSCertPem == nil {
		slog.Warn("HTTPS key/cert not set; falling back to auto self-signed")
		certFile := filepath.Join(serverruntime.DataDir, "cert.pem")
		keyFile := filepath.Join(serverruntime.DataDir, "key.pem")
		if err := serverruntime.EnsureSelfSignedCert(certFile, keyFile); err != nil {
			slog.Error("failed to ensure self-signed cert", "error", err)
		}
		restapi.StartHttpsServer(certFile, keyFile, false)
		return
	}
	restapi.StartHttpsServerFromStrings(*cxsettingspkg.CxSettings.HTTPSCertPem, *cxsettingspkg.CxSettings.HTTPSKeyPem, false)
}

func restartNetworkListenersFromSettings() {
	slog.Info("Syncing network listeners with settings")
	_ = restapi.ShutdownHttpServerWithTimeout(time.Second*5, true)
	_ = restapi.ShutdownHttpsServerWithTimeout(time.Second*5, true)
	time.Sleep(300 * time.Millisecond)
	startConfiguredListeners(currentListenerConfig(), "Restarting")
}
