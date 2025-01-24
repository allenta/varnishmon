package assets

import "embed"

//go:embed all:static
var StaticFS embed.FS

//go:embed all:templates
var TemplatesFS embed.FS
