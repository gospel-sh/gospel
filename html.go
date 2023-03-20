package gospel

import (
	"fmt"
)

type Element interface {
	RenderElement(Context) string
}

type Attribute interface {
	RenderAttribute(Context) string
}

type HTMLElement struct {
	Tag        string
	Void       bool
	Children   []Element
	Attributes []Attribute
}

type HTMLAttribute struct {
	Name  string
	Value any
	Args  []any
}

func Attrib(tag string) func(value any, args ...any) *HTMLAttribute {
	return func(value any, args ...any) *HTMLAttribute {
		return &HTMLAttribute{
			Name:  tag,
			Value: value,
			Args:  args,
		}
	}
}

func (a *HTMLAttribute) RenderAttribute(c Context) string {

	extraArgs := ""

	if len(a.Args) > 0 {

		for _, arg := range a.Args {
			if strArg, ok := arg.(string); ok {
				extraArgs += " " + strArg
			}
		}

	}

	return fmt.Sprintf("%s=\"%s%s\"", a.Name, a.Value, extraArgs)
}

func (h *HTMLElement) RenderElement(c Context) string {

	renderedAttributes := ""

	for _, attribute := range h.Attributes {
		renderedAttributes += " " + attribute.RenderAttribute(c)
	}

	if h.Void {
		return fmt.Sprintf("<%[1]s%[2]s/>", h.Tag, renderedAttributes)
	} else {

		renderedChildren := ""

		for _, child := range h.Children {
			renderedChildren += child.RenderElement(c)
		}

		return fmt.Sprintf("<%[1]s%[3]s>%[2]s</%[1]s>", h.Tag, renderedChildren, renderedAttributes)
	}

}

type Literal struct {
	Value string
}

func (l *Literal) RenderElement(c Context) string {
	return l.Value
}

func children(args ...any) (children []Element) {

	children = make([]Element, 0, len(args))

	for _, arg := range args {
		if elem, ok := arg.(Element); ok {
			children = append(children, elem)
		} else if str, ok := arg.(string); ok {
			children = append(children, &Literal{str})
		}
	}

	return
}

func attributes(args ...any) (attributes []Attribute) {

	attributes = make([]Attribute, 0, len(args))

	for _, arg := range args {
		if elem, ok := arg.(Attribute); ok {
			attributes = append(attributes, elem)
		}
	}

	return
}

type Fragment struct {
	Children []Element
}

func (f *Fragment) RenderElement(c Context) string {
	renderedElements := ""

	for _, element := range f.Children {
		renderedElements += element.RenderElement(c)
	}

	return renderedElements
}

func F(args ...any) Element {
	return &Fragment{
		Children: children(args...),
	}
}

func Tag(tag string) func(args ...any) Element {
	return func(args ...any) Element {
		return &HTMLElement{tag, false, children(args...), attributes(args...)}
	}
}

func VoidTag(tag string) func(args ...any) Element {
	return func(args ...any) Element {
		return &HTMLElement{tag, true, children(args...), attributes(args...)}
	}
}

// HTML Attributes

var Lang = Attrib("lang")
var Charset = Attrib("charset")
var Rel = Attrib("rel")
var Sizes = Attrib("sizes")
var Href = Attrib("href")
var Type = Attrib("type")
var Class = Attrib("class")
var Id = Attrib("id")
var Src = Attrib("src")
var OnClick = Attrib("onClick")
var OnChange = Attrib("onChange")
var OnSubmit = Attrib("onSubmit")
var Name = Attrib("name")
var Value = Attrib("value")
var Style = Attrib("style")

// HTML Tags
// https://www.w3.org/TR/2011/WD-html-markup-20110113/syntax.html

var Html = Tag("html")
var Div = Tag("div")
var Title = Tag("title")
var Head = Tag("head")
var Body = Tag("body")
var StyleTag = Tag("style")
var Address = Tag("address")
var Aside = Tag("aside")
var Header = Tag("header")
var Footer = Tag("footer")
var H1 = Tag("h1")
var H2 = Tag("h2")
var H3 = Tag("h3")
var H4 = Tag("h4")
var H5 = Tag("h5")
var H6 = Tag("h6")
var Main = Tag("main")
var Nav = Tag("nav")
var Section = Tag("section")
var Blockquote = Tag("blockquote")
var Dd = Tag("dd")
var Dl = Tag("dl")
var Dt = Tag("dt")
var Figcaption = Tag("figcaption")
var Figure = Tag("figure")
var Li = Tag("li")
var Menu = Tag("menu")
var Ol = Tag("ol")
var P = Tag("p")
var Pre = Tag("pre")
var Ul = Tag("ul")
var A = Tag("a")
var Abbr = Tag("abbr")
var B = Tag("b")
var Bdi = Tag("bdi")
var Bdo = Tag("bdo")
var Cite = Tag("cite")
var Code = Tag("code")
var Data = Tag("data")
var Dfn = Tag("dfn")
var Em = Tag("em")
var I = Tag("i")
var Kbd = Tag("kbd")
var Mark = Tag("mark")
var G = Tag("g")
var Rp = Tag("rp")
var Rt = Tag("rt")
var Ruby = Tag("ruby")
var S = Tag("s")
var Samp = Tag("samp")
var Small = Tag("small")
var Span = Tag("span")
var Strong = Tag("strong")
var Sub = Tag("sub")
var Sup = Tag("sup")
var Time = Tag("time")
var U = Tag("u")
var Var = Tag("var")
var Map = Tag("map")
var Video = Tag("video")
var Iframe = Tag("iframe")
var Object = Tag("object")
var Picture = Tag("picture")
var Portal = Tag("portal")
var Svg = Tag("svg")
var Math = Tag("math")
var Canvas = Tag("canvas")
var Noscript = Tag("noscript")
var Script = Tag("script")
var Del = Tag("del")
var Ins = Tag("ins")
var Caption = Tag("caption")
var Colgroup = Tag("colgroup")
var Table = Tag("table")
var Tbody = Tag("tbody")
var Td = Tag("td")
var Tfoot = Tag("tfoot")
var Th = Tag("th")
var Thead = Tag("thead")
var Tr = Tag("tr")
var Button = Tag("button")
var Datalist = Tag("datalist")
var Fieldset = Tag("fieldset")
var Form = Tag("form")
var Label = Tag("label")
var Legend = Tag("legend")
var Meter = Tag("meter")
var Optgroup = Tag("optgroup")
var Option = Tag("option")
var Output = Tag("output")
var Progress = Tag("progress")
var Select = Tag("select")
var Textarea = Tag("textarea")
var Details = Tag("details")
var Dialog = Tag("dialog")
var Summary = Tag("summary")
var Slot = Tag("slot")
var Template = Tag("template")

// HTML Void Tags

var Area = VoidTag("area")
var Base = VoidTag("base")
var Br = VoidTag("br")
var Col = VoidTag("col")
var Command = VoidTag("command")
var Embed = VoidTag("embed")
var Hr = VoidTag("hr")
var Img = VoidTag("img")
var Input = VoidTag("input")
var Keygen = VoidTag("keygen")
var Link = VoidTag("link")
var Meta = VoidTag("meta")
var Param = VoidTag("param")
var Source = VoidTag("source")
var Track = VoidTag("track")
var Wbr = VoidTag("wbr")
