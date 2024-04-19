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

func ApplyAny[T any](applyArgs ...func(args ...any) T) func(args ...any) []any {
	return func(args ...any) []any {
		rules := make([]any, 0, len(applyArgs))
		for _, arg := range applyArgs {
			rules = append(rules, arg(args...))
		}
		return rules
	}
}
