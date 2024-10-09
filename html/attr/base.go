package attr

import (
	"html"
	"html/template"
)

type Attribute template.HTMLAttr

func (a Attribute) String() string { return string(a) }

func Attr(key, value string) Attribute {
	return Attribute(key + `="` + html.EscapeString(value) + `"`)
}
