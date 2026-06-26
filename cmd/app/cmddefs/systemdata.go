package cmddefs

import (
	"chronix/cmd/app/rpc"
	"chronix/cmd/app/systemdata"
	serverruntime "chronix/internal/serverruntime"
	"fmt"

	u "github.com/bcicen/go-units"
)

type (
	SystemDataCommandDef struct {
		Systemdata SystemDataCommand `cmd:"" help:"Show the system data"`
	}
	SystemDataCommand struct{}
)

func init() {
	if err := rpc.RegisterName("SystemData", &SystemDataCommand{}); err != nil {
		panic(err)
	}
}

func (f *SystemDataCommand) Run() error {
	var s string
	err := rpc.Call("SystemData.GetSystemData", &struct{}{}, &s)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func (f *SystemDataCommand) GetSystemData(_ *struct{}, data *string) error {
	opts := u.FmtOptions{
		Label:     true,
		Short:     true,
		Precision: 2,
	}
	sd := systemdata.GetSystemData()

	*data = fmt.Sprintf(systemDataFormat,
		u.NewValue(float64(sd.Alloc), u.Byte).MustConvert(u.MegaByte).Fmt(opts),
		u.NewValue(float64(sd.SystemAlloc), u.Byte).MustConvert(u.MegaByte).Fmt(opts),
		sd.NumGoRoutines,
		sd.NumCPUs,
		sd.CPUPercent,
		serverruntime.DataDir)
	return nil
}

const systemDataFormat = `
Alloc: %s
SystemAlloc: %s
NumGoRoutines: %d
NumCPUs: %d
CPUPercent: %.1f
DataDir: %s
`
