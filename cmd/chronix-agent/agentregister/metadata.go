package agentregister

import (
	"bufio"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

func getMacOSVersionName(v string) string {
	if strings.HasPrefix(v, "26.") {
		return "Tahoe"
	}
	if strings.HasPrefix(v, "15.") {
		return "Sequoia"
	}
	if strings.HasPrefix(v, "14.") {
		return "Sonoma"
	}
	if strings.HasPrefix(v, "13.") {
		return "Ventura"
	}
	if strings.HasPrefix(v, "12.") {
		return "Monterey"
	}
	if strings.HasPrefix(v, "11.") {
		return "Big Sur"
	}
	return ""
}

func CollectMetadata() map[string]any {
	meta := map[string]any{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpu":  runtime.NumCPU(),
	}
	if u, err := user.Current(); err == nil {
		meta["user"] = u.Username
	}

	osType := runtime.GOOS
	osVersion := ""

	switch runtime.GOOS {
	case "linux":
		if f, err := os.Open("/etc/os-release"); err == nil {
			defer func() { _ = f.Close() }()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "ID=") {
					osType = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
				} else if strings.HasPrefix(line, "VERSION_ID=") {
					osVersion = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
				}
			}
		}
	case "darwin":
		osType = "macos"
		if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			osVersion = strings.TrimSpace(string(out))
			name := getMacOSVersionName(osVersion)
			if name != "" {
				osVersion = name + " " + osVersion
			}
		}
	case "windows":
		osType = "windows"
		cmdText := "$v = Get-ItemProperty 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion'; $cap = (Get-CimInstance Win32_OperatingSystem).Caption; $dv = $v.DisplayVersion; if (!$dv) { $dv = $v.ReleaseId }; $cap + ' ' + $dv + ' (Build ' + $v.CurrentBuild + ')'"
		if out, err := exec.Command("powershell", "-Command", cmdText).Output(); err == nil {
			osVersion = strings.TrimSpace(string(out))
		}
	}

	meta["os_type"] = osType
	meta["os_version"] = osVersion
	return meta
}
