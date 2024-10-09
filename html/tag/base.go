package tag

import (
	"fmt"
	"html"
	"strings"

	"github.com/emad-elsaid/go-server/html/attr"
)

type Element string

func (e Element) String() string { return string(e) }

func Tag(name string, cs ...fmt.Stringer) Element {
	var b strings.Builder

	b.WriteString("<" + name)

	for i := range cs {
		if _, ok := cs[i].(attr.Attribute); ok {
			b.WriteRune(' ')
			b.WriteString(cs[i].String())
		}
	}

	b.WriteRune('>')

	for i := range cs {
		if _, ok := cs[i].(Element); ok {
			b.WriteString(cs[i].String())
		}
	}

	b.WriteString("</" + name + ">")

	return Element(b.String())
}

func HTML(c ...fmt.Stringer) Element { return Tag("html", c...) }

func String(s string) Element {
	return Element(html.EscapeString(s))
}
