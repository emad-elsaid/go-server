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

func String(s string) Element { return Element(html.EscapeString(s)) }
