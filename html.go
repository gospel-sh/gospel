// Gospel - Golang Simple Extensible Web Framework
// Copyright (C) 2019-2024 - The Gospel Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

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
	RenderElement() string
}

type Attribute interface {
	RenderAttribute() string
}

type HTMLElement struct {
	Tag        string                 `json:"tag"`
	Void       bool                   `json:"void"`
	Value      any                    `json:"value"`
	Safe       bool                   `json:"safe"`
	Children   []any                  `json:"children" graph:"include"`
	Attributes []*HTMLAttribute       `json:"attributes" graph:"include"`
	Args       []any                  `json:"args" graph:"ignore"`
	Decorators []HTMLElementDecorator `json:"-"`
}

func (h *HTMLElement) RenderCodeChildren() string {
	renderedChildren := ""

	for _, child := range h.Children {

		if child == nil {
			continue
		}

		// we first check if this is a generator
		if generator, ok := child.(Generator); ok {
			renderedChildren += generator.RenderCode()
			continue
		}

		htmlChild, ok := child.(*HTMLElement)

		if !ok {

			htmlFuncChild, ok := child.(PureElementFunction)

			if ok {

				htmlChild, ok = htmlFuncChild().(*HTMLElement)

				if !ok {
					continue
				}
			} else {
				continue
			}

		}

		// the child element can still be nil
		if htmlChild == nil {
			continue
		}

		renderedChildren += htmlChild.RenderCode()
	}

	return renderedChildren
}

func (a *HTMLAttribute) RenderCodeAttribute() string {

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

	if a.Value == nil {
		return a.Name
	}

	strValue, ok := a.Value.(string)

	if !ok {
		return ""
	}

	return fmt.Sprintf("%s=\"%s%s\"", a.Name, strValue, extraArgs)
}

func (h *HTMLElement) RenderCode() string {

	if strValue, ok := h.Value.(string); ok {
		// this is a literal element
		return strValue
	}

	renderedAttributes := ""

	for _, attribute := range h.Attributes {

		ra := attribute.RenderCodeAttribute()

		if ra == "" {
			continue
		}

		renderedAttributes += " " + ra
	}

	if h.Void {
		return fmt.Sprintf("<%[1]s%[2]s/>", h.Tag, renderedAttributes)
	} else {

		renderedChildren := h.RenderCodeChildren()

		if h.Tag == "" {
			return renderedChildren
		}

		return fmt.Sprintf("<%[1]s%[3]s>%[2]s</%[1]s>", h.Tag, renderedChildren, renderedAttributes)
	}
}

func (h *HTMLElement) Copy() *HTMLElement {

	newChildren := make([]any, len(h.Children))
	newArgs := make([]any, len(h.Args))
	newAttributes := make([]*HTMLAttribute, len(h.Attributes))
	newDecorators := make([]HTMLElementDecorator, len(h.Decorators))

	copy(newChildren, h.Children)
	copy(newArgs, h.Args)
	copy(newAttributes, h.Attributes)
	copy(newDecorators, h.Decorators)

	return &HTMLElement{
		Tag:        h.Tag,
		Void:       h.Void,
		Value:      h.Value,
		Safe:       h.Safe,
		Children:   newChildren,
		Attributes: newAttributes,
		Decorators: newDecorators,
		Args:       newArgs,
	}
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

func DataAttrib(name, value string) *HTMLAttribute {

	dataName := Fmt("data-%s", name)

	if value == "" {
		return &HTMLAttribute{
			Name:  dataName,
			Value: nil,
		}
	}

	return &HTMLAttribute{
		Name:  dataName,
		Value: value,
	}
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

func (a *HTMLAttribute) RenderAttribute() string {

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

		if stringer, ok := a.Value.(fmt.Stringer); !ok {
			return ""
		} else {
			strValue = stringer.String()
		}
	}

	return fmt.Sprintf("%s=\"%s%s\"", html.EscapeString(a.Name), html.EscapeString(strValue), extraArgs)
}

func (h *HTMLElement) RenderElement() string {

	if strValue, ok := h.Value.(string); ok {

		if h.Safe {
			return strValue
		}

		// this is a literal element
		return html.EscapeString(strValue)
	}

	renderedAttributes := ""

	for _, attribute := range h.Attributes {

		ra := attribute.RenderAttribute()

		if ra == "" {
			continue
		}

		renderedAttributes += " " + ra
	}

	if h.Void {
		return fmt.Sprintf("<%[1]s%[2]s/>", h.Tag, renderedAttributes)
	} else {

		renderedChildren := h.RenderChildren()

		if h.Tag == "" {
			return renderedChildren
		}

		return fmt.Sprintf("<%[1]s%[3]s>%[2]s</%[1]s>", h.Tag, renderedChildren, renderedAttributes)
	}

}

func (h *HTMLElement) RenderChildren() string {
	renderedChildren := ""

	for _, child := range h.Children {

		if child == nil {
			continue
		}

		htmlChild, ok := child.(*HTMLElement)

		if !ok {

			htmlFuncChild, ok := child.(PureElementFunction)

			if ok {

				htmlChild, ok = htmlFuncChild().(*HTMLElement)

				if !ok {
					continue
				}
			} else {
				// to do: how to handle this?
				Log.Warning("Could not render child of type %T", child)
				continue
			}

		}

		// the child element can still be nil
		if htmlChild == nil {
			continue
		}

		renderedChildren += htmlChild.RenderElement()
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

func children(args ...any) (chldr []any) {

	chldr = make([]any, 0, len(args))

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
		} else if elem, ok := arg.(*HTMLElement); ok && elem != nil {
			chldr = append(chldr, elem)
		} else if str, ok := arg.(string); ok {
			chldr = append(chldr, Literal(str))
		} else if _, ok := arg.(PureElementFunction); ok {
			chldr = append(chldr, arg)
		} else if _, ok := arg.(Generator); ok {
			chldr = append(chldr, arg)
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

					return []*HTMLAttribute{
						&HTMLAttribute{
							Name:  "name",
							Value: selectedValue.ScopedId(),
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

			htmlChild, ok := child.(*HTMLElement)

			if !ok {
				continue
			}

			if htmlChild.Tag != "option" {
				continue
			}

			var value any

			for _, attrib := range htmlChild.Attributes {

				if attrib.Name == "value" {
					value = attrib.Value
					break
				}
			}

			if value == selectedValue.GetRaw() {
				htmlChild.Attributes = append(htmlChild.Attributes, BooleanAttrib("selected")())
			}
		}
	}
}

func (h *HTMLElement) Generate(c Context) (any, error) {
	newChildren := make([]any, 0, len(h.Children))
	// we only iterate over the args
	for _, child := range h.Children {
		if ga, ok := child.(Generator); ok {
			if element, err := ga.Generate(c); err != nil {
				return nil, fmt.Errorf("cannot generate argument: %v", err)
			} else {
				if generatedChildren, ok := element.([]any); ok {
					newChildren = append(newChildren, generatedChildren...)
				} else {
					newChildren = append(newChildren, element)
				}
			}
		} else {
			newChildren = append(newChildren, child)
		}
	}

	el := h.Copy()
	el.Children = newChildren

	// we apply all decorators to the element
	for _, decorator := range el.Decorators {
		decorator(el)
	}

	return el, nil
}

func (h *HTMLElement) Attribute(name string) *HTMLAttribute {
	for _, attribute := range h.Attributes {
		if attribute.Name == name {
			return attribute
		}
	}
	return nil
}

func Assignable(asChild bool) HTMLElementDecorator {
	return func(element *HTMLElement) {

		assignableMapper := func(htmlAttrib *HTMLAttribute) []*HTMLAttribute {

			if htmlAttrib.Name == "value" {

				var v ContextVarObj
				var dv any
				var ok bool

				v, ok = htmlAttrib.Value.(ContextVarObj)

				if ok {

					if len(htmlAttrib.Args) == 1 {
						// there was a default value passed in
						dv = htmlAttrib.Args[0]
					} else {
						// we get the raw value instead
						dv = v.GetRaw()
					}
				} else {
					// this is a regular attribute
					return []*HTMLAttribute{htmlAttrib}
				}

				htmlAttrib.Name = "gospel-value"
				htmlAttrib.Hidden = true

				if asChild {

					strValue, ok := dv.(string)

					if !ok {
						// to do: add a warning
						return []*HTMLAttribute{htmlAttrib}
					}

					element.Children = append(element.Children, Literal(strValue))

					return []*HTMLAttribute{htmlAttrib, &HTMLAttribute{
						Name:  "name",
						Value: v.ScopedId(),
					}}

				}

				return []*HTMLAttribute{htmlAttrib, &HTMLAttribute{
					Name:  "value",
					Value: dv,
				}, &HTMLAttribute{
					Name:  "name",
					Value: v.ScopedId(),
				}}
			}

			return []*HTMLAttribute{htmlAttrib}
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

		valueMapper := func(htmlAttrib *HTMLAttribute) []*HTMLAttribute {

			if htmlAttrib.Name == "gospel-value" {

				v, ok := htmlAttrib.Value.(ContextVarObj)

				if !ok {
					Log.Warning("uh oh: not a context var")
					return nil
				}

				if values, ok := form[v.ScopedId()]; ok && len(values) > 0 {
					value := values[0]
					v.Set(value)
				}

			}
			return nil
		}

		mapHTMLAttributes(htmlChild.Attributes, valueMapper)

		// we recurse into child elements...
		assignVars(c, form, htmlChild)

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

		var formData *FormData
		var method = "GET"
		var hasFormData bool

		for _, arg := range element.Args {
			if formData, hasFormData = arg.(*FormData); hasFormData {
				break
			}
		}

		for _, attrib := range element.Attributes {
			if attrib.Name == "method" {
				if strValue, ok := attrib.Value.(string); !ok {
					Log.Warning("%v form method is not a string", element)
				} else {
					// to do: check method type
					method = strValue
				}
			}
		}

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
				var id string

				if idAttr := element.Attribute("id"); idAttr != nil {
					id = Cast(idAttr.Value, f.Id())
				} else {
					id = f.Id()
				}

				if req.Method == method && c.Interactive() {

					if HasContentType(req, "multipart/form-data") {
						if err := req.ParseMultipartForm(1024 * 1024 * 10); err != nil {
							Log.Error("Cannot parse form: %v", err)
							return nil
						}
					} else if err := req.ParseForm(); err != nil {
						Log.Error("Cannot parse form: %v", err)
						return nil
					}

					if req.Form.Get("_gspl") == id {

						if formData != nil {
							// we set the form data, if it is defined
							formData.Set(req.Form)
						}

						// we update variables based on form content
						assignVars(c, req.Form, element)

						// we call the function

						f.Call()

					}

				}

				// we append the ID of the form
				element.Children = append(element.Children, Input(Type("hidden"), Name("_gspl"), Value(id)))

				return []*HTMLAttribute{&HTMLAttribute{
					Name:  "gospel-onSubmit",
					Value: id,
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
		Decorators: decorators,
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

		htmlChild, ok := child.(*HTMLElement)

		if !ok {
			continue
		}

		if htmlChild.Value != "" {
			htmlChild.Safe = true
		}
	}
}

type ElementsMap map[string]func(args ...any) *HTMLElement

var elements ElementsMap = ElementsMap{}

func Tag(tag string, decorators ...HTMLElementDecorator) func(args ...any) *HTMLElement {
	elements[tag] = func(args ...any) *HTMLElement {
		return makeTag(tag, args, false, decorators)
	}
	return elements[tag]
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
var GospelValue = Attrib("gospel-value")
var Style = Attrib("style")
var Method = Attrib("method")
var Content = Attrib("content")
var Alt = Attrib("alt")
var As = Attrib("as")
var Enctype = Attrib("enctype")
var Placeholder = Attrib("placeholder")
var Defer = BooleanAttrib("defer")
var Download = BooleanAttrib("download")
var Aria = func(tag string, value any, args ...any) *HTMLAttribute {
	return Attrib(tag)(value, args...)
}

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
var EmTag = Tag("em")
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
var SubTag = Tag("sub")
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
var Button = Tag("button", Assignable(false))
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
var LiteralTag = Tag("")

// Safe values
var Nbsp = SafeLiteral("&nbsp;")
var Gt = SafeLiteral("&gt;")
var L = func(literal string) *HTMLElement {
	return SafeLiteral(literal)
}

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

// Tree walking

func Walk[T any](element any, walker func(t T, element *HTMLElement)) {

	walk := func(value any, element *HTMLElement) {
		if t, ok := value.(T); ok {
			walker(t, element)
		}
	}

	htmlElement, ok := element.(*HTMLElement)

	if !ok {
		if htmlElementFunc, ok := element.(PureElementFunction); ok {
			if htmlElement, ok = htmlElementFunc().(*HTMLElement); !ok {
				return
			} else if htmlElement == nil {
				return
			}
		}
		return
	}

	if htmlElement == nil {
		return
	}

	walk(htmlElement, htmlElement)

	for _, attribute := range htmlElement.Attributes {
		walk(attribute, htmlElement)
	}

	for _, arg := range htmlElement.Args {
		if t, ok := arg.(T); ok {
			walker(t, htmlElement)
		}
	}

	for _, child := range htmlElement.Children {
		Walk(child, walker)
	}
}
