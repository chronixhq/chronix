// Package chronix provides embedded documentation assets for the application.
package chronix

import "embed"

// HelpMarkdown contains the embedded help documentation from docs/help.md.
//
//go:embed docs/help.md
var HelpMarkdown embed.FS
