package attr

import (
	"html"
)

type Attribute string

func (a Attribute) String() string { return string(a) }

func Attr(key, value string) Attribute {
	return Attribute(key + `="` + html.EscapeString(value) + `"`)
}
