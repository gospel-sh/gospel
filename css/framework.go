package css

import (
	"github.com/gospel-sh/gospel"
)

var DefaultConfig = Config{
	Borders{},
	Padding{},
}

var CSS = MakeCSS(DefaultConfig)

type Config struct {
	Borders Borders
	Padding Padding
}

type Borders struct{}
type Padding struct{}

func MakeCSS(config Config) *gospel.Stylesheet {
	css := gospel.MakeStylesheet()

	MakeBorders(config.Borders, css)

	return css
}

func MakeBorders(config Borders, css *gospel.Stylesheet) {

}

func MakePadding(config Padding, css *gospel.Stylesheet) {

}
