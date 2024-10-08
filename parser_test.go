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

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

var source = `
html template test <p buz=bam>
	<p foo=bar>
		this is some text
	</p>
	this is some more text
	<div
		style="text-decoration: none"
		class="foo bar baz"
	>

	</div>
</p>

html template foo <p></p>
`

var expectedProgram = &Program{
	Statements: []*Statement{
		{
			Type: HTMLStmt,
			HTML: &HTMLStatement{
				Type: TemplateStmt,
				Template: &HTMLTemplate{
					Name: "test",
					Element: &HTMLElement{
						Tag: "p",
						Attributes: []*HTMLAttribute{
							{
								Name:  "buz",
								Value: "bam",
							},
						},
						Children: []any{
							&HTMLElement{
								Tag: "p",
								Children: []any{
									&HTMLElement{
										Value: "this is some text",
										Safe:  false,
									},
								},
								Attributes: []*HTMLAttribute{
									{
										Name:  "foo",
										Value: "bar",
									},
								},
							},
							&HTMLElement{
								Value: "this is some more text",
								Safe:  false,
							},
							&HTMLElement{
								Tag:      "div",
								Children: []any{},
								Attributes: []*HTMLAttribute{
									{
										Name:  "style",
										Value: "text-decoration: none",
									},
									{
										Name:  "class",
										Value: "foo bar baz",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Type: HTMLStmt,
			HTML: &HTMLStatement{
				Type: TemplateStmt,
				Template: &HTMLTemplate{
					Name: "foo",
					Element: &HTMLElement{
						Children:   []any{},
						Attributes: []*HTMLAttribute{},
						Tag:        "p",
					},
				},
			},
		},
	},
}

func TestParser(t *testing.T) {

	parser := &Parser{}
	program, err := parser.Parse(source)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(program, expectedProgram); diff != "" {
		t.Fatalf("invalid result: %s", diff)
	}

}

func BenchmarkParser(b *testing.B) {
	parser := &Parser{}
	b.SetBytes(int64(len(source)))

	for i := 0; i < b.N; i++ {
		if _, err := parser.Parse(source); err != nil {
			b.Fatal(err)
		}
	}

}
