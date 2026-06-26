package chronixreactapp

import (
	"embed"
	"io/fs"
)

// DistFS exposes the embedded React build output for the server runtime.
//
//go:embed dist
var DistFS embed.FS

func DistSubFS() (fs.FS, error) {
	return fs.Sub(DistFS, "dist")
}
