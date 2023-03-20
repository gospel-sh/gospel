package gospel

import (
	"io/fs"
)

type App struct {
	Root         func(Context) Element
	StaticFiles  []fs.FS
	StaticPrefix string
}
