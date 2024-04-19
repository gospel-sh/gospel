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

package gospel

// Generates an element or other renderable "thing"
type Generator interface {
	// generate a renderable element from the value
	Generate(c Context) (any, error)
	// render the source code of the generator
	RenderCode() string
}

type GeneratorFunction = func(c Context) (any, error)

// Generates a HTML element
type HTMLElementGenerator struct {
	HTMLElement
}

// Generates a HTML attribute
type HTMLAttributeGenerator struct {
	HTMLAttribute
}

func (h *HTMLAttributeGenerator) Generate(c Context) (any, error) {
	return nil, nil
}
