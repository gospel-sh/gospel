# Gospel CSS

## Interface

```golang

import (
	. "github.com/gospel-sh/gospel/css"
)

var CSS = StyleSheet()

// we define our radius as a variable
var Radius = Px(2)

var Rounded = CSS.Rule(
	BorderRadius(Radius),
	BorderWidth("2px"),
	BorderStyle("solid"),
)

var Dark = Class("Dark")

var PrimaryColor = RGB("#fff")
var MainFontFamily = String("Arial, Helvetica")

var Text = CSS.Rule(
	// declaration
	Color(PrimaryColor),
	// declaration
	FontFamily(MainFontFamily),
	// declaration
	FontSize("12px"),
)

var Button = CSS.Rule(
	// rule
	Rounded,
	// declaration
	BackgroundColor("#333"),
	// rule
	Dark.Rule(
		// declaration
		BackgroundColor("#eee"),
	),
	// rule
	Mobile(
		// declaration
		Width("100%"),
	),
)

```

```css

.dark .button {
	background: #eee;
}

.button {
	background: #333;
}

@media(mobile) {
	.button {
		width: 100%;
	}
}

```