package gospel

// Generates an element or other renderable "thing"
type Generator interface {
	Generate(c Context) (any, error)
}

type GeneratorFunction = func(c Context) (any, error)

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
