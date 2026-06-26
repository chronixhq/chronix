package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/alecthomas/kong"
)

var CLI struct {
	DataDir string `help:"Directory used for tester state files." env:"CHRONIX_TESTER_DATA_DIR" type:"path"`

	Run struct {
		ResultsPort      int           `help:"Port for the results UI and summary API." default:"5180"`
		APIPort          int           `help:"Port for the API fixture service." default:"5181"`
		WebhookPort      int           `help:"Port for the webhook capture service." default:"5182"`
		IMAPPollInterval time.Duration `help:"Interval between IMAP inbox polls." default:"30s"`
	} `cmd:"" default:"1" help:"Start the tester services."`

	Config struct {
		Imap struct {
			Host string `required:"" help:"IMAP host."`
			Port int    `default:"993" help:"IMAP port."`
			User string `required:"" help:"IMAP user."`
			Pass string `required:"" help:"IMAP password."`
			SSL  bool   `default:"true" help:"Use SSL when connecting."`
		} `cmd:"" help:"Configure IMAP polling for notification verification."`
	} `cmd:"" help:"Configure tester settings."`

	GenerateToken struct {
	} `cmd:"" help:"Generate a JSON token for shell variable-capture testing."`

	Bootstrap struct {
		ChronixDB       string `arg:"" help:"Path to the Chronix SQLite database."`
		TesterHost      string `arg:"" help:"Host name or IP address used by Chronix to reach the tester services."`
		APIPort         int    `help:"Port used by the tester API base URL." default:"5181"`
		WebhookPort     int    `help:"Port used by the tester webhook base URL." default:"5182"`
		TokenCommand    string `help:"Override the shell command used by the bootstrapped token-generation step."`
		ShellWorkingDir string `help:"Override the working directory used by the bootstrapped shell command." type:"path"`
	} `cmd:"" help:"Bootstrap a Chronix database with tester fixtures."`

	Reset struct {
	} `cmd:"" help:"Delete and recreate the tester state databases."`

	Version struct {
	} `cmd:"" help:"Print version information."`
}

func main() {
	initLogger()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			printVersion()
			return
		}
	}

	ctx := kong.Parse(
		&CLI,
		kong.Name("chronix-tester"),
		kong.Description("Chronix fixture harness for local execution, captures, and release testing."),
	)

	switch ctx.Command() {
	case "run":
		exitOnError(runServices())
	case "config imap":
		exitOnError(saveImapConfig())
	case "generate-token":
		exitOnError(generateToken())
	case "bootstrap <chronix-db> <tester-host>":
		exitOnError(runBootstrap())
	case "reset":
		exitOnError(resetState())
	case "version":
		printVersion()
	default:
		slog.Error("unknown command", "command", ctx.Command())
		os.Exit(1)
	}
}

func initLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
}

func exitOnError(err error) {
	if err == nil {
		return
	}
	slog.Error("chronix tester failed", "error", err)
	os.Exit(1)
}
