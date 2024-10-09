package main

import (
	a "github.com/emad-elsaid/go-server/html/attr"
	t "github.com/emad-elsaid/go-server/html/tag"
)

func main() {
	Get("/", func(r Request) Response {
		return Text(
			Layout(
				t.String("Hello World!"),
			),
		)
	})

	Start()
}

func Layout(view t.Element) t.Element {
	return t.Html(
		a.Lang("en"),
		t.Head(
			t.Meta(a.Charset("utf-8")),
			t.Meta(a.Name("viewport"), a.Content("width=device-width, initial-scale=1")),
			t.Link(a.Rel("stylesheet"), a.Href("/public/style.css?v="+Sha256("public/style.css"))),
			t.Title(t.String("Hello World!")),
		),
		t.Body(
			t.Section(a.Class("section"),
				t.Div(a.Class("container is-max-desktop"),
					view,
				),
			),
		),
	)
}
