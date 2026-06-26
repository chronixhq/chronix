package app

import (
	"chronix/cmd/app/rpc"
	"chronix/internal/execution"
	jobrunpkg "chronix/internal/jobrun"
	progresspkg "chronix/internal/progress"
	serverruntime "chronix/internal/serverruntime"
	"chronix/internal/updater"
	"chronix/pkg/fileutil"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dan-sherwin/go-app-settings"
)

func init() {
	switch runtime.GOOS {
	case "linux":
		if os.Getuid() == 0 {
			serverruntime.DataDir = "/var/lib/chronix"
		} else {
			dataHome := os.Getenv("XDG_DATA_HOME")
			if dataHome != "" {
				if info, err := os.Stat(dataHome); err == nil && info.IsDir() {
					if fileutil.IsOwnedByUser(info) {
						serverruntime.DataDir = filepath.Join(dataHome, "chronix")
					}
				}
			}
			if serverruntime.DataDir == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					serverruntime.DataDir = "/tmp/chronix"
				} else {
					serverruntime.DataDir = filepath.Join(homeDir, ".local", "share", "chronix")
				}
			}
		}
	case "darwin":
		if os.Getuid() == 0 {
			serverruntime.DataDir = "/Library/Application Support/Chronix"
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			serverruntime.DataDir = filepath.Join(homeDir, "Library", "Application Support", "Chronix")
		}
	case "windows":
		serverruntime.DataDir = filepath.Join("C:\\ProgramData", "Chronix")
	}

	var cliDataDirConfig struct {
		DataDir string `short:"D" long:"datadir" help:"Path to data directory" default:"${data_dir}" env:"CHRONIX_DATA_DIR"`
	}
	vars := kong.Vars{
		"data_dir": serverruntime.DataDir,
	}
	cliParser, err := kong.New(&cliDataDirConfig, kong.Name("chronix"), kong.Exit(nil), kong.Help(nil), vars)
	if err == nil {
		var filteredArgs []string
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "-D" || os.Args[i] == "--datadir" {
				filteredArgs = append(filteredArgs, os.Args[i])
				if i+1 < len(os.Args) {
					filteredArgs = append(filteredArgs, os.Args[i+1])
					i++
				}
			}
		}
		_, err := cliParser.Parse(filteredArgs)
		if err == nil && cliDataDirConfig.DataDir != "" {
			serverruntime.DataDir = cliDataDirConfig.DataDir
		}
	}

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			if s == "" {
				return fmt.Errorf("log level cannot be empty")
			}
			LoggingLevel = s
			initLogger()
			return nil
		},
		GetFunc:     func() string { return LoggingLevel },
		Name:        "log_level",
		Description: "Logging level (debug|info|warn|error)",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			if s == "" {
				return fmt.Errorf("rpc socket path cannot be empty")
			}
			rpc.SocketPath = s
			return nil
		},
		GetFunc:     func() string { return rpc.SocketPath },
		Name:        "rpc_socket_path",
		Description: "Path to Unix domain socket for RPC",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("invalid jobqueue enqueue burst: %v", err)
			}
			if n < 1 {
				return fmt.Errorf("jobqueue enqueue burst must be >= 1")
			}
			jobrunpkg.SetJobQueueEnqueueBurst(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(jobrunpkg.GetJobQueueEnqueueBurst()) },
		Name:        "jobqueue_enqueue_burst",
		Description: "Job queue ingress burst size for 1/ms rate limiter",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("invalid jobqueue worker count: %v", err)
			}
			if n < 1 {
				return fmt.Errorf("jobqueue worker count must be >= 1")
			}
			jobrunpkg.SetWorkerCount(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(jobrunpkg.GetWorkerCount()) },
		Name:        "jobqueue_worker_count",
		Description: "Number of concurrent job workers",
	})

	var execStepTimeout = 30 * time.Second
	var execRetryCount = 1
	var execRowCap = 100
	var progressBufCap = 200

	applyExecDefaults := func() {
		execution.SetExecutorDefaults(execStepTimeout, execRetryCount, execRowCap)
		progresspkg.SetConfig(progressBufCap)
	}

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("invalid job step timeout: %v", err)
			}
			execStepTimeout = d
			applyExecDefaults()
			return nil
		},
		GetFunc:     func() string { return execStepTimeout.String() },
		Name:        "job_step_timeout",
		Description: "Default per-step timeout",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 0 {
				return fmt.Errorf("invalid job step retry count")
			}
			execRetryCount = n
			applyExecDefaults()
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(execRetryCount) },
		Name:        "job_step_retry_count",
		Description: "Number of retries per step (in addition to first attempt)",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			if s == "" {
				return fmt.Errorf("updater manifest URL cannot be empty")
			}
			updater.ManifestURL = s
			return nil
		},
		GetFunc:     func() string { return updater.ManifestURL },
		Name:        "updater_manifest_url",
		Description: "URL to fetch the latest version information",
	})
	app_settings.RegisterDurationSetting("updater_check_interval", "Interval between automatic update checks", &updater.CheckInterval)

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n <= 0 {
				return fmt.Errorf("invalid job result row cap")
			}
			execRowCap = n
			applyExecDefaults()
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(execRowCap) },
		Name:        "job_result_row_cap",
		Description: "Max rows to capture for SQLQuery steps",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n <= 0 {
				return fmt.Errorf("invalid job progress buffer size")
			}
			progressBufCap = n
			applyExecDefaults()
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(progressBufCap) },
		Name:        "job_progress_buffer_size",
		Description: "Per-run progress event buffer size",
	})
}
