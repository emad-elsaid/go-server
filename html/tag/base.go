package tag

import (
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/emad-elsaid/go-server/html/attr"
)

type Element template.HTML

func (e Element) String() string { return string(e) }

func Tag(name string, cs ...fmt.Stringer) Element {
	var b strings.Builder

	b.WriteRune('<')
	b.WriteString(name)

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

	b.WriteString("</")
	b.WriteString(name)
	b.WriteRune('>')

	return Element(b.String())
}

func Tags(cs ...Element) Element {
	var b strings.Builder

	for i := range cs {
		b.WriteString(cs[i].String())
	}

	return Element(b.String())
}

func VoidTag(name string, cs ...attr.Attribute) Element {
	var b strings.Builder

	b.WriteString("<" + name)

	for i := range cs {
		b.WriteRune(' ')
		b.WriteString(cs[i].String())
	}

	b.WriteString("/>")

	return Element(b.String())
}

func String(s string) Element { return Element(html.EscapeString(s)) }
