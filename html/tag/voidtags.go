package tag
import "github.com/emad-elsaid/go-server/html/attr"
func Area(c ...attr.Attribute) Element { return VoidTag("area", c...) }
func Base(c ...attr.Attribute) Element { return VoidTag("base", c...) }
func Br(c ...attr.Attribute) Element { return VoidTag("br", c...) }
func Col(c ...attr.Attribute) Element { return VoidTag("col", c...) }
func Embed(c ...attr.Attribute) Element { return VoidTag("embed", c...) }
func Hr(c ...attr.Attribute) Element { return VoidTag("hr", c...) }
func Img(c ...attr.Attribute) Element { return VoidTag("img", c...) }
func Input(c ...attr.Attribute) Element { return VoidTag("input", c...) }
func Link(c ...attr.Attribute) Element { return VoidTag("link", c...) }
func Meta(c ...attr.Attribute) Element { return VoidTag("meta", c...) }
func Param(c ...attr.Attribute) Element { return VoidTag("param", c...) }
func Source(c ...attr.Attribute) Element { return VoidTag("source", c...) }
func Track(c ...attr.Attribute) Element { return VoidTag("track", c...) }
func Wbr(c ...attr.Attribute) Element { return VoidTag("wbr", c...) }