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

package main

import (
	"embed"
	. "github.com/gospel-sh/gospel"
	"github.com/gospel-sh/gospel/examples"
	"io/fs"
	"os"
	"os/signal"
)

//go:embed static
var StaticFiles embed.FS
var StaticFilesPrefix, _ = fs.Sub(StaticFiles, "static")

// Serves the Gospel examples
func makeExamples() *Server {

	root := func(c Context) Element {

		router := UseRouter(c)

		return F(
			Doctype("html"),
			Html(
				Lang("en"),
				Head(
					Meta(Charset("utf-8")),
					Title("Gospel Examples"),
					Script(Defer(), Src("/static/gospel.js"), Type("module")),
				),
				Body(
					router.Match(
						c,
						Route("/css", examples.CSSExample),
						Route("", Div(
							H1("Gospel Examples"),
							Ul(
								Li(
									A(Href("/css"), "CSS example"),
								),
							),
						)),
					),
				),
			),
		)

	}

	return MakeServer(&App{
		Root:         root,
		StaticFiles:  []fs.FS{StaticFilesPrefix},
		StaticPrefix: "/static",
	})
}

func main() {
	examplesServer := makeExamples()
	examplesServer.Start()

	Log.Info("Server running...")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// we wait for an interrupt...
	<-c

	Log.Info("Stopping server...")

	examplesServer.Stop()

}
