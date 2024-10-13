package main

import (
	b "github.com/willoma/bulma-gomponents"
	x "maragu.dev/gomponents-htmx"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	Get("/{$}", func(r Request) Response {
		return Ok(
			Layout(
				Navbar(),
				b.Section(Text("Hello World!")),
			),
		)
	})

	Get("/about", func(r Request) Response {
		return Ok(
			Layout(
				Navbar(),
				b.Section(Text("About")),
			),
		)
	})

	Start()
}

func Navbar() Node {
	return b.Navbar(
		b.Dark,
		x.Boost("true"),
		b.NavbarStart(
			b.NavbarAHref("/", "Home"),
			b.NavbarAHref("/about", "About"),
		),
	)
}

func Layout(view ...Node) Node {
	return b.HTML(
		Lang("en"),
		b.HTitle("Hello World!"),
		b.Stylesheet("/public/style.css?v="+Sha256("public/style.css")),
		b.Script("https://unpkg.com/htmx.org@2.0.3"),
		Group(view),
	)
}
