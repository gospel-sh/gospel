package gospel

import (
	"encoding/hex"
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
	Name   string
	Hidden bool
	Value  any
	Args   []any
}

func Unhex(value string) []byte {
	if v, err := hex.DecodeString(value); err != nil {
		Log.Error("Cannot parse hex value: %v", err)
		return nil
	} else {
		return v
	}
}

func Hex(value []byte) string {
	return hex.EncodeToString(value)
}

func Fmt(text string, args ...any) string {
	return fmt.Sprintf(text, args...)
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

	if a.Hidden {
		return ""
	}

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

		ra := attribute.RenderAttribute(c)

		if ra == "" {
			continue
		}

		renderedAttributes += " " + ra
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

func children(args ...any) (chldr []Element) {

	chldr = make([]Element, 0, len(args))

	for _, arg := range args {
		if elementList, ok := arg.([]Element); ok {
			chldr = append(chldr, elementList...)
		} else if anyList, ok := arg.([]any); ok {
			chldr = append(chldr, children(anyList...)...)
		} else if elem, ok := arg.(Element); ok {
			chldr = append(chldr, elem)
		} else if str, ok := arg.(string); ok {
			chldr = append(chldr, &Literal{str})
		}
	}

	return
}

func attributes(args ...any) (attribs []Attribute) {

	attribs = make([]Attribute, 0, len(args))

	for _, arg := range args {
		if elem, ok := arg.(Attribute); ok {
			attribs = append(attribs, elem)
		} else if attribList, ok := arg.([]Attribute); ok {
			attribs = append(attribs, attribList...)
		} else if anyList, ok := arg.([]any); ok {
			attribs = append(attribs, attributes(anyList...)...)
		}
	}

	return
}

type Fragment struct {
	Children []Element
}

type HTMLElementDecorator func(*HTMLElement)

func mapHTMLAttributes(attribs []Attribute, mapper func(*HTMLAttribute) []Attribute) []Attribute {
	newAttribs := make([]Attribute, 0, len(attribs))

	for _, attrib := range attribs {
		if htmlAttrib, ok := attrib.(*HTMLAttribute); ok {
			newAttribs = append(newAttribs, mapper(htmlAttrib)...)
		}
	}

	return newAttribs

}

func Assignable() HTMLElementDecorator {
	return func(element *HTMLElement) {

		assignableMapper := func(htmlAttrib *HTMLAttribute) []Attribute {

			if htmlAttrib.Name == "value" {
				v, ok := htmlAttrib.Value.(ContextVarObj)
				if !ok {
					Log.Warning("uh oh: not a HTML attribute")
					return nil
				}

				htmlAttrib.Name = "gospel-value"
				htmlAttrib.Hidden = true

				return []Attribute{htmlAttrib, &HTMLAttribute{
					Name:  "value",
					Value: v.GetRaw(),
				}, &HTMLAttribute{
					Name:  "name",
					Value: v.Id(),
				}}
			}

			return []Attribute{htmlAttrib}
		}

		element.Attributes = mapHTMLAttributes(element.Attributes, assignableMapper)

	}
}

func assignVars(c Context, form map[string][]string, element *HTMLElement) {
	for _, child := range element.Children {
		htmlChild, ok := child.(*HTMLElement)

		if !ok {
			continue
		}

		valueMapper := func(htmlAttrib *HTMLAttribute) []Attribute {

			Log.Info("Name: %s", htmlAttrib.Name)

			if htmlAttrib.Name == "gospel-value" {
				v, ok := htmlAttrib.Value.(ContextVarObj)

				if !ok {
					Log.Warning("uh oh: not a context var")
					return nil
				}

				Log.Info("Variable: %s", v.Id())

				if values, ok := form[v.Id()]; ok && len(values) > 0 {
					value := values[0]
					v.Set(value)
				}

			}
			return nil
		}

		mapHTMLAttributes(htmlChild.Attributes, valueMapper)

		// we recurse into child elements...
		assignVars(c, form, htmlChild)

		Log.Info("%s", htmlChild.Tag)
	}
}

func Submittable() HTMLElementDecorator {
	// Check if there's an OnSubmit attribute with a callback
	// If given, add some JS code to ensure we call it
	return func(element *HTMLElement) {

		submitMapper := func(htmlAttrib *HTMLAttribute) []Attribute {
			if htmlAttrib.Name == "onSubmit" {
				f, ok := htmlAttrib.Value.(ContextFuncObj)
				if !ok {
					Log.Warning("uh oh: not a function")
					return nil
				}

				c := f.Context()

				router := UseRouter(c)

				req := router.Request()
				Log.Info("%s:%s", req.URL.Path, req.Method)

				if req.Method == "POST" && c.Interactive() {

					if err := req.ParseForm(); err != nil {
						Log.Error("Cannot parse form: %v", err)
						return nil
					}

					// we update variables based on form content
					assignVars(c, req.Form, element)

					// we call the function

					f.Call()

					Log.Info("%v", req.Form)

				}

				return []Attribute{&HTMLAttribute{
					Name:  "gospel-onSubmit",
					Value: f.Id(),
				}}
			}

			return []Attribute{htmlAttrib}
		}

		element.Attributes = mapHTMLAttributes(element.Attributes, submitMapper)
	}
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

func makeTag(tag string, args []any, void bool, decorators []any) Element {

	element := &HTMLElement{tag, void, children(args...), attributes(args...)}

	// we apply all decorators to the element
	for _, decorator := range decorators {
		if df, ok := decorator.(HTMLElementDecorator); ok {
			df(element)
		}
	}

	return element

}

func Tag(tag string, decorators ...any) func(args ...any) Element {
	return func(args ...any) Element {
		return makeTag(tag, args, false, decorators)
	}
}

func VoidTag(tag string, decorators ...any) func(args ...any) Element {
	return func(args ...any) Element {
		return makeTag(tag, args, true, decorators)
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
var Method = Attrib("method")
var Content = Attrib("content")
var Alt = Attrib("alt")
var Placeholder = Attrib("placeholder")

// HTML Tags
// https://www.w3.org/TR/2011/WD-html-markup-20110113/syntax.html

var Html = Tag("html")
var Div = Tag("div")
var Title = Tag("title")
var Head = Tag("head")
var Body = Tag("body")
var StyleTag = Tag("style") // renamed as it clashes with 'Style' helper
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
var VarTag = Tag("var") // renamed as it clashes with 'Var'
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
var Form = Tag("form", Submittable())
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
var Input = VoidTag("input", Assignable())
var Keygen = VoidTag("keygen")
var Link = VoidTag("link")
var Meta = VoidTag("meta")
var Param = VoidTag("param")
var Source = VoidTag("source")
var Track = VoidTag("track")
var Wbr = VoidTag("wbr")

// Special tags
var Doctype = func(doctype string) *Literal { return &Literal{fmt.Sprintf("<!doctype %s>", doctype)} }
