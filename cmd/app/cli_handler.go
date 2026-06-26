package app

import (
	cmddefspkg "chronix/cmd/app/cmddefs"
	"chronix/cmd/app/consts"
	cxrestapi "chronix/cxrestapi"
	serverruntime "chronix/internal/serverruntime"
	"chronix/pkg/buildinfo"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/dan-sherwin/go-app-settings"
	"github.com/willabides/kongplete"
)

type (
	CLIConfig struct {
		Datadir     string                   `short:"D" long:"datadir" help:"Path to data directory" default:"${data_dir}" env:"CHRONIX_DATA_DIR"`
		Conf        string                   `short:"c" long:"conf" help:"Path to configuration file (reserved for future use)" hidden:""`
		VersionFlag kong.VersionFlag         `name:"version" short:"v" help:"Print version information and quit"`
		AppSettings app_settings.SettingsDef `embed:""`
		Commands
		Run                RunCommand                   `cmd:"" help:"Run application in foreground" default:"1" hidden:""`
		InstallCompletions kongplete.InstallCompletions `cmd:"" name:"completionscript" help:"Install shell completions (bash|zsh|fish)." hidden:""`
	}

	RestrictedCLIConfig struct {
		Datadir     string           `short:"D" long:"datadir" help:"Path to data directory" default:"${data_dir}" env:"CHRONIX_DATA_DIR"`
		Conf        string           `short:"c" long:"conf" help:"Path to configuration file (reserved for future use)" hidden:""`
		VersionFlag kong.VersionFlag `name:"version" short:"v" help:"Print version information and quit"`
		cxrestapi.HTTPCommandDef
		cmddefspkg.VersionCommandDef
		Run RunCommand `cmd:"" help:"Run application in foreground" default:"1" hidden:""`
	}
)

var (
	CLICommand *kong.Context
	cliConfig  CLIConfig
	vars       = kong.Vars{}
)

func processCLI(dbExists bool) {
	vars["logging_level"] = LoggingLevel
	vars["data_dir"] = serverruntime.DataDir
	vars["version"] = buildinfo.Version

	var cli interface{} = &cliConfig
	var restrictedConfig RestrictedCLIConfig
	if !dbExists {
		cli = &restrictedConfig
	}

	parser, err := kong.New(cli,
		kong.Name(consts.APPNAME),
		kong.Description(consts.APPNAME+" application"),
		kong.ShortUsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		vars,
	)
	if err != nil {
		panic(err)
	}

	kongplete.Complete(parser)
	CLICommand, err = parser.Parse(os.Args[1:])
	if err != nil {
		if !dbExists {
			fmt.Println("Error: Application needs to be initialized first. Run application in foreground first to initialize it.")
			os.Exit(1)
		}
		parser.FatalIfErrorf(err)
	}
	if !dbExists {
		cxrestapi.CLIHTTPConfig = restrictedConfig.HTTPCommandDef
	} else {
		cxrestapi.CLIHTTPConfig = cliConfig.HTTPCommandDef
	}
	postParseProcessing(CLICommand, dbExists, &cliConfig, &restrictedConfig)
}
