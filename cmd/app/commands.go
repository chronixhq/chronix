package app

import (
	cmddefspkg "chronix/cmd/app/cmddefs"
	cxrestapi "chronix/cxrestapi"
	serverruntime "chronix/internal/serverruntime"

	"github.com/alecthomas/kong"
)

type (
	Commands struct {
		cmddefspkg.BuildInfoCommandDef
		cmddefspkg.VersionCommandDef
		cmddefspkg.AgentsCommandDef
		cmddefspkg.SystemDataCommandDef
		cxrestapi.AuthCommandDef
		serverruntime.ServerCommandDef
		cxrestapi.HTTPCommandDef
		cmddefspkg.UpdateCommandDef
	}
)

func postParseProcessing(_ *kong.Context, _ bool, _ *CLIConfig, _ *RestrictedCLIConfig) {
}
