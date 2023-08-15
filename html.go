package gospel

import (
	"encoding/hex"
	"fmt"
	"html"
	"mime"
	"net/http"
	"net/url"
	"strings"
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
	Value      any
	Safe       bool
	Children   []*HTMLElement
	Attributes []*HTMLAttribute
	Args       []any
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

func BooleanAttrib(tag string) func() *HTMLAttribute {
	return func() *HTMLAttribute {
		return &HTMLAttribute{
			Name: tag,
		}
	}
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
				extraArgs += " " + html.EscapeString(strArg)
			}
		}

	}

	if a.Value == nil {
		return fmt.Sprintf("%s", html.EscapeString(a.Name))
	}

	strValue, ok := a.Value.(string)

	if !ok {
		return ""
	}

	return fmt.Sprintf("%s=\"%s%s\"", html.EscapeString(a.Name), html.EscapeString(strValue), extraArgs)
}

func (h *HTMLElement) RenderElement(c Context) string {

	if strValue, ok := h.Value.(string); ok {

		if h.Safe {
			return strValue
		}

		// this is a literal element
		return html.EscapeString(strValue)
	}

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

		renderedChildren := h.RenderChildren(c)

		if h.Tag == "" {
			return renderedChildren
		}

		return fmt.Sprintf("<%[1]s%[3]s>%[2]s</%[1]s>", h.Tag, renderedChildren, renderedAttributes)
	}

}

func (h *HTMLElement) RenderChildren(c Context) string {
	renderedChildren := ""

	for _, child := range h.Children {

		if child == nil {
			continue
		}

		renderedChildren += child.RenderElement(c)
	}

	return renderedChildren
}

func SafeLiteral(value string) *HTMLElement {
	return &HTMLElement{
		Value: value,
		Safe:  true,
	}
}

func Literal(value string) *HTMLElement {
	return &HTMLElement{
		Value: value,
		Safe:  false,
	}
}

func children(args ...any) (chldr []*HTMLElement) {

	chldr = make([]*HTMLElement, 0, len(args))

	for _, arg := range args {
		if elementList, ok := arg.([]Element); ok {
			for _, elem := range elementList {
				chldr = append(chldr, children(elem)...)
			}
		} else if htmlList, ok := arg.([]*HTMLElement); ok {
			for _, elem := range htmlList {
				chldr = append(chldr, children(elem)...)
			}
		} else if anyList, ok := arg.([]any); ok {
			chldr = append(chldr, children(anyList...)...)
		} else if elem, ok := arg.(*HTMLElement); ok {
			chldr = append(chldr, elem)
		} else if str, ok := arg.(string); ok {
			chldr = append(chldr, Literal(str))
		}
	}

	return
}

func attributes(args ...any) (attribs []*HTMLAttribute) {

	attribs = make([]*HTMLAttribute, 0, len(args))

	for _, arg := range args {

		if elem, ok := arg.(*HTMLAttribute); ok {

			if elem == nil {
				continue
			}

			attribs = append(attribs, elem)
		} else if attribList, ok := arg.([]Attribute); ok {
			attribs = append(attribs, attributes(attribList)...)
		} else if anyList, ok := arg.([]any); ok {
			attribs = append(attribs, attributes(anyList...)...)
		}
	}

	return
}

type HTMLElementDecorator func(*HTMLElement)

func mapHTMLAttributes(attribs []*HTMLAttribute, mapper func(*HTMLAttribute) []*HTMLAttribute) []*HTMLAttribute {
	newAttribs := make([]*HTMLAttribute, 0, len(attribs))

	for _, attrib := range attribs {
		newAttribs = append(newAttribs, mapper(attrib)...)
	}

	return newAttribs

}

func Selectable() HTMLElementDecorator {

	return func(element *HTMLElement) {

		var selectedValue ContextVarObj

		mapper := func(htmlAttrib *HTMLAttribute) []*HTMLAttribute {

			if htmlAttrib.Name == "value" {

				var ok bool

				selectedValue, ok = htmlAttrib.Value.(ContextVarObj)

				if ok {
					return []*HTMLAttribute{&HTMLAttribute{
						Name:  "name",
						Value: selectedValue.Id(),
					},
						&HTMLAttribute{
							Name:   "gospel-value",
							Hidden: true,
							Value:  selectedValue,
						},
					}
				}

			}

			return []*HTMLAttribute{htmlAttrib}

		}

		element.Attributes = mapHTMLAttributes(element.Attributes, mapper)

		if selectedValue == nil {
			return
		}

		// we have found a value, we map it to the children

		for _, child := range element.Children {

			if child.Tag != "option" {
				continue
			}

			var value any

			for _, attrib := range child.Attributes {

				if attrib.Name == "value" {
					value = attrib.Value
					break
				}
			}

			if value == selectedValue.GetRaw() {
				child.Attributes = append(child.Attributes, BooleanAttrib("selected")())
			}
		}
	}
}

func Assignable(asChild bool) HTMLElementDecorator {
	return func(element *HTMLElement) {

		assignableMapper := func(htmlAttrib *HTMLAttribute) []*HTMLAttribute {

			if htmlAttrib.Name == "value" {
				v, ok := htmlAttrib.Value.(ContextVarObj)

				if !ok {
					// this is a regular attribute
					return []*HTMLAttribute{htmlAttrib}
				}

				htmlAttrib.Name = "gospel-value"
				htmlAttrib.Hidden = true

				if asChild {

					strValue, ok := v.GetRaw().(string)

					if !ok {
						// to do: add a warning
						return []*HTMLAttribute{htmlAttrib}
					}

					element.Children = append(element.Children, Literal(strValue))

					return []*HTMLAttribute{htmlAttrib, &HTMLAttribute{
						Name:  "name",
						Value: v.Id(),
					}}

				}

				return []*HTMLAttribute{htmlAttrib, &HTMLAttribute{
					Name:  "value",
					Value: v.GetRaw(),
				}, &HTMLAttribute{
					Name:  "name",
					Value: v.Id(),
				}}
			}

			return []*HTMLAttribute{htmlAttrib}
		}

		element.Attributes = mapHTMLAttributes(element.Attributes, assignableMapper)

	}
}

func assignVars(c Context, form map[string][]string, element *HTMLElement) {
	for _, child := range element.Children {

		valueMapper := func(htmlAttrib *HTMLAttribute) []*HTMLAttribute {

			if htmlAttrib.Name == "gospel-value" {
				v, ok := htmlAttrib.Value.(ContextVarObj)

				if !ok {
					Log.Warning("uh oh: not a context var")
					return nil
				}

				if values, ok := form[v.Id()]; ok && len(values) > 0 {
					value := values[0]
					v.Set(value)
				}

			}
			return nil
		}

		mapHTMLAttributes(child.Attributes, valueMapper)

		// we recurse into child elements...
		assignVars(c, form, child)

	}
}

// Determine whether the request `content-type` includes a
// server-acceptable mime-type
//
// Failure should yield an HTTP 415 (`http.StatusUnsupportedMediaType`)
func HasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}

type FormValues url.Values

func Submittable() HTMLElementDecorator {
	// Check if there's an OnSubmit attribute with a callback
	// If given, add some J Scode to ensure we call it
	return func(element *HTMLElement) {

		submitMapper := func(htmlAttrib *HTMLAttribute) []*HTMLAttribute {
			if htmlAttrib.Name == "onSubmit" {
				f, ok := htmlAttrib.Value.(ContextFuncObj[any])
				if !ok {
					Log.Warning("uh oh: not a function")
					return nil
				}

				c := f.Context()

				router := UseRouter(c)
				req := router.Request()

				if req.Method == "POST" && c.Interactive() {

					if HasContentType(req, "multipart/form-data") {
						if err := req.ParseMultipartForm(1024 * 1024 * 10); err != nil {
							Log.Error("Cannot parse form: %v", err)
							return nil
						}
					} else if err := req.ParseForm(); err != nil {
						Log.Error("Cannot parse form: %v", err)
						return nil
					}

					if req.Form.Get("_gospel_id") == f.Id() {

						// we update variables based on form content
						assignVars(c, req.Form, element)

						// we call the function

						f.Call()

					}

				}

				// we append the ID of the form
				element.Children = append(element.Children, Input(Type("hidden"), Name("_gospel_id"), Value(f.Id())))

				return []*HTMLAttribute{&HTMLAttribute{
					Name:  "gospel-onSubmit",
					Value: f.Id(),
				}}
			}

			return []*HTMLAttribute{htmlAttrib}
		}

		element.Attributes = mapHTMLAttributes(element.Attributes, submitMapper)
	}
}

func F(args ...any) *HTMLElement {
	return &HTMLElement{
		Children: children(args...),
	}
}

func makeTag(tag string, args []any, void bool, decorators []HTMLElementDecorator) *HTMLElement {

	element := &HTMLElement{
		Tag:        tag,
		Void:       void,
		Children:   children(args...),
		Attributes: attributes(args...),
		Args:       args,
	}

	// we apply all decorators to the element
	for _, decorator := range decorators {
		decorator(element)
	}

	return element

}

// Marks children of a script node as safe so that we don't escape special characters...
func isScript(element *HTMLElement) {
	for _, child := range element.Children {
		if child.Value != "" {
			child.Safe = true
		}
	}
}

func Tag(tag string, decorators ...HTMLElementDecorator) func(args ...any) *HTMLElement {
	return func(args ...any) *HTMLElement {
		return makeTag(tag, args, false, decorators)
	}
}

func VoidTag(tag string, decorators ...HTMLElementDecorator) func(args ...any) *HTMLElement {
	return func(args ...any) *HTMLElement {
		return makeTag(tag, args, true, decorators)
	}
}

// HTML Attributes

var Role = Attrib("role")
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
var As = Attrib("as")
var Enctype = Attrib("enctype")
var Placeholder = Attrib("placeholder")
var Defer = BooleanAttrib("defer")

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
var Script = Tag("script", isScript)
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
var Textarea = Tag("textarea", Assignable(true))
var Details = Tag("details")
var Dialog = Tag("dialog")
var Summary = Tag("summary")
var Slot = Tag("slot")
var Template = Tag("template")
var Select = Tag("select", Selectable())

// Safe values
var Nbsp = SafeLiteral("&nbsp;")

// HTML Void Tags

var Area = VoidTag("area")
var Base = VoidTag("base")
var Br = VoidTag("br")
var Col = VoidTag("col")
var Command = VoidTag("command")
var Embed = VoidTag("embed")
var Hr = VoidTag("hr")
var Img = VoidTag("img")
var Input = VoidTag("input", Assignable(false))
var Keygen = VoidTag("keygen")
var Link = VoidTag("link")
var Meta = VoidTag("meta")
var Param = VoidTag("param")
var Source = VoidTag("source")
var Track = VoidTag("track")
var Wbr = VoidTag("wbr")

// Special tags
var Doctype = func(doctype string) *HTMLElement {
	return &HTMLElement{Safe: true, Value: fmt.Sprintf("<!doctype %s>", doctype)}
}
