package main

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func main() {
	Get("/", func(r Request) Response {
		return TextO(
			Layout(
				Text("Hello World!"),
			),
		)
	})

	Start()
}

func Metas() Node {
	return Group{
		Meta(Charset("utf-8")),
		Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
	}
}

func Layout(view Node) Node {
	return HTML(
		Lang("en"),
		Head(
			Metas(),
			Link(Rel("stylesheet"), Href("/public/style.css?v="+Sha256("public/style.css"))),
			Title("Hello World!"),
		),
		Body(
			Section(Class("section"),
				Div(Class("container is-max-desktop"),
					view,
				),
			),
		),
	)
}
