package agentconfig

import (
	"chronix/pkg/fileutil"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ServerHost    string `json:"serverHost"`
	WSPort        int    `json:"wsPort"`
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	PrivKeyB64    string `json:"privKey"`
	PubKeyB64     string `json:"pubKey"`
	ServerSPKIB64 string `json:"serverSpki,omitempty"`
}

func DefaultPath() string {
	var dir string
	switch runtime.GOOS {
	case "linux":
		if os.Getuid() == 0 {
			dir = "/var/lib/chronix-agent"
		} else {
			dataHome := os.Getenv("XDG_DATA_HOME")
			if dataHome != "" {
				if info, err := os.Stat(dataHome); err == nil && info.IsDir() {
					if fileutil.IsOwnedByUser(info) {
						dir = filepath.Join(dataHome, "chronix-agent")
					}
				}
			}
			if dir == "" {
				homeDir, _ := os.UserHomeDir()
				if homeDir != "" {
					dir = filepath.Join(homeDir, ".local", "share", "chronix-agent")
				} else {
					dir = "."
				}
			}
		}
	case "darwin":
		if os.Getuid() == 0 {
			dir = "/Library/Application Support/ChronixAgent"
		} else {
			homeDir, _ := os.UserHomeDir()
			if homeDir != "" {
				dir = filepath.Join(homeDir, "Library", "Application Support", "ChronixAgent")
			} else {
				dir = "."
			}
		}
	case "windows":
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		dir = filepath.Join(programData, "ChronixAgent")
	default:
		dir = "."
	}
	if env := os.Getenv("CHRONIX_AGENT_DATA_DIR"); env != "" {
		dir = env
	}
	return filepath.Join(dir, "agent.json")
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()
	return os.Rename(tmp, path)
}
