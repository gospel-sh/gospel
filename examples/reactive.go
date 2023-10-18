package examples

import (
	. "github.com/gospel-sh/gospel"
)

type RLiteral[T any] struct {
	Value T
}

func (r *RLiteral[T]) Equals(o RLiteral[T]) RLiteral[bool] {
	return RLiteral[bool]{false}
}

func (r *RLiteral[T]) Toggle() any {
	return nil
}

type RList[T any] struct {
	Items []T
}

func (r *RList[T]) Map(mapper func(T) Element) []Element {
	mappedElements := make([]Element, 0, len(r.Items))

	for _, item := range r.Items {
		mappedElements = append(mappedElements, mapper(item))
	}

	return mappedElements
}

func (r *RList[T]) Any(filter func(T) RLiteral[bool]) RLiteral[bool] {
	return RLiteral[bool]{false}
}

func (r *RList[T]) Filter(filter func(T) RLiteral[bool]) *RList[T] {
	newList := &RList[T]{
		Items: make([]T, 0, len(r.Items)),
	}
	return newList
}

type Item struct {
	Title       RLiteral[string]
	Description RLiteral[string]
	Category    RLiteral[string]
	Active      RLiteral[bool]
	Children    RList[*Item]
}

type Category struct {
	Name     RLiteral[string]
	Selected RLiteral[bool]
}

func (c *Category) IsSelected([]*Category) bool {
	return false
}

// the individual category values will be detected and identified, they will
// be passed to JS so that we can reference them there...
func filters(categories []*Category) Element {

	items := make([]Element, 0, len(categories))

	for _, category := range categories {
		items = append(items, Li(category.Name, OnClick(category.Selected.Toggle())))
	}

	return F(
		items,
	)
}

func And(conditions ...RLiteral[bool]) RLiteral[bool] {
	return RLiteral[bool]{false}
}

// there will be
func ItemList(items RList[*Item], categories RList[*Category]) Element {
	return F(
		items.Filter(func(item *Item) RLiteral[bool] {
			// we only display items that are in the selected categories
			return categories.Any(func(c *Category) RLiteral[bool] { return And(c.Selected, c.Name.Equals(item.Category)) })
		}).Map(
			func(item *Item) Element {
				// we toggle the active status of an item
				return Li(item.Title, OnClick(item.Active.Toggle()))
			},
		),
	)
}
