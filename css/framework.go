// Gospel - Golang Simple Extensible Web Framework
// Copyright (C) 2019-2024 - The Gospel Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

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

	css := gospel.MakeStylesheet("gospel")

	MakeBorders(config.Borders, css)

	return css
}

func MakeBorders(config Borders, css *gospel.Stylesheet) {

}

func MakePadding(config Padding, css *gospel.Stylesheet) {

}
