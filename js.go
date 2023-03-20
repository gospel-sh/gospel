package gospel

import (
	"embed"
	"io/fs"
)

//go:embed js/dist
var JSRaw embed.FS

var JS, _ = fs.Sub(JSRaw, "js/dist")
