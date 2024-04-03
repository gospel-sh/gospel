package gospel

import (
	"fmt"
)

// Generates an element or other renderable "thing"
type Generator interface {
	Generate(c Context) (any, error)
}

// Generates a HTML element
type HTMLElementGenerator struct {
	HTMLElement
}

// Generates a HTML attribute
type HTMLAttributeGenerator struct {
	HTMLAttribute
}

func (h *HTMLAttributeGenerator) Generate(c Context) (any, error) {
	return nil, nil
}

func (h *HTMLElementGenerator) Generate(c Context) (any, error) {

	newArgs := make([]any, 0, len(h.Args))
	// we only iterate over the args
	for _, arg := range h.Args {
		if ga, ok := arg.(Generator); ok {
			if element, err := ga.Generate(c); err != nil {
				return nil, fmt.Errorf("cannot generate argument: %v", err)
			} else {
				newArgs = append(newArgs, element)
			}
		} else {
			newArgs = append(newArgs, arg)
		}
	}

	if ef, ok := elements[h.Tag]; !ok {
		return nil, fmt.Errorf("unknown HTML tag: %s", h.Tag)
	} else {
		return ef(newArgs...), nil
	}
}

func init() {
	h := &HTMLElementGenerator{HTMLElement{Tag: "html", Safe: false, Value: "buh"}}
	var he any = h

	if hg, ok := he.(Generator); ok {
		fmt.Println("ok")
		if v, err := hg.Generate(&DefaultContext{}); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("sucess: %v", v)
		}

	} else {
		fmt.Println("not ok")
	}
}
