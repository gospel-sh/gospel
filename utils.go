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

func If[T any](condition bool, value T) T {
	if condition {
		return value
	}
	return *new(T)
}

func IfElse[T any](condition bool, value T, alternative T) T {
	if condition {
		return value
	} else {
		return alternative
	}
}

func DoIf[T any](condition bool, value func() T) T {
	if condition {
		return value()
	}
	return *new(T)
}

func Cast[T any](v any, d T) T {
	if vt, ok := v.(T); ok {
		return vt
	}
	return d
}
