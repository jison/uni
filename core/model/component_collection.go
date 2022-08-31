package model

import "fmt"

type ComponentIterator interface {
	Iterate(f func(Component) bool) bool
}

type ComponentSlice []Component

type ComponentCollection interface {
	ComponentIterator
	Each(f func(Component))
	Filter(predicate func(Component) bool) ComponentCollection
	ToArray() ComponentSlice
	ToSet() ComponentSet
	Distinct() ComponentCollection
}

type ComponentSet interface {
	ComponentCollection
	Contains(Component) bool
	Len() int
}

type FuncComponentIterator func(f func(Component) bool) bool

func (i FuncComponentIterator) Iterate(f func(Component) bool) bool {
	return i(f)
}

type filteredComponentIterator struct {
	ComponentIterator
	predicate func(Component) bool
}

func (i *filteredComponentIterator) Iterate(f func(Component) bool) bool {
	return i.ComponentIterator.Iterate(func(c Component) bool {
		if i.predicate(c) {
			return f(c)
		}
		return true
	})
}

type combinedComponentIterator struct {
	left  ComponentIterator
	right ComponentIterator
}

func (i *combinedComponentIterator) Iterate(f func(Component) bool) bool {
	isContinue := i.left.Iterate(f)
	if !isContinue {
		return false
	}

	return i.right.Iterate(f)
}

type distinctComponentIterator struct {
	original ComponentIterator
}

func (i *distinctComponentIterator) Iterate(f func(Component) bool) bool {
	haveMet := map[Component]struct{}{}
	return i.original.Iterate(func(c Component) bool {
		if _, ok := haveMet[c]; ok {
			return true
		}
		haveMet[c] = struct{}{}
		return f(c)
	})
}

var _ ComponentCollection = ComponentSlice{}

func (cs ComponentSlice) Iterate(f func(Component) bool) bool {
	for _, c := range cs {
		if !f(c) {
			return false
		}
	}
	return true
}

func (cs ComponentSlice) Each(f func(Component)) {
	_componentsEach(cs, f)
}

func (cs ComponentSlice) Filter(predicate func(Component) bool) ComponentCollection {
	return _componentsFilter(cs, predicate)
}

func (cs ComponentSlice) ToArray() ComponentSlice {
	return cs
}

func (cs ComponentSlice) ToSet() ComponentSet {
	return _componentsToSet(cs)
}

func (cs ComponentSlice) Distinct() ComponentCollection {
	return _componentsDistinct(cs)
}

func (cs ComponentSlice) Format(fs fmt.State, r rune) {
	_componentsFormat(cs, fs, r)
}

func _componentsEach(it ComponentIterator, f func(Component)) {
	it.Iterate(func(c Component) bool {
		f(c)
		return true
	})
}

func _componentsFilter(it ComponentIterator, predicate func(Component) bool) ComponentCollection {
	return ComponentsOfIterator(&filteredComponentIterator{it, predicate})
}

func _componentsToArray(it ComponentIterator) ComponentSlice {
	var arr []Component
	it.Iterate(func(c Component) bool {
		arr = append(arr, c)
		return true
	})

	return arr
}

func _componentsToSet(it ComponentIterator) ComponentSet {
	set := newComponentSet()
	it.Iterate(func(c Component) bool {
		set.Add(c)
		return true
	})
	return set
}

func _componentsDistinct(it ComponentIterator) ComponentCollection {
	return ComponentsOfIterator(&distinctComponentIterator{it})
}

func _componentsFormat(it ComponentIterator, s fmt.State, r rune) {
	_, _ = fmt.Fprint(s, "[")
	var comFormat string
	if s.Flag('#') && r == 'v' {
		comFormat = "%#v"
	} else if s.Flag('+') && r == 'v' {
		comFormat = "%+v"
	} else {
		comFormat = "%v"
	}

	first := true
	it.Iterate(func(c Component) bool {
		if !first {
			_, _ = fmt.Fprint(s, ", ")
		} else {
			first = false
		}

		_, _ = fmt.Fprintf(s, comFormat, c)
		return true
	})
	_, _ = fmt.Fprint(s, "]")
}

type componentCollection struct {
	ComponentIterator
}

var _ ComponentCollection = &componentCollection{}

func (s *componentCollection) Each(f func(Component)) { _componentsEach(s, f) }
func (s *componentCollection) Filter(predicate func(Component) bool) ComponentCollection {
	return _componentsFilter(s, predicate)
}
func (s *componentCollection) ToArray() ComponentSlice       { return _componentsToArray(s) }
func (s *componentCollection) ToSet() ComponentSet           { return _componentsToSet(s) }
func (s *componentCollection) Distinct() ComponentCollection { return _componentsDistinct(s) }
func (s *componentCollection) Format(fs fmt.State, r rune)   { _componentsFormat(s, fs, r) }

func ComponentsOfIterator(it ComponentIterator) ComponentCollection {
	return &componentCollection{it}
}

type emptyComponentCollection struct{}

var _ ComponentSet = emptyComponentCollection{}

func (s emptyComponentCollection) Iterate(_ func(Component) bool) bool               { return true }
func (s emptyComponentCollection) Each(_ func(Component))                            {}
func (s emptyComponentCollection) Filter(_ func(Component) bool) ComponentCollection { return s }
func (s emptyComponentCollection) ToArray() ComponentSlice                           { return []Component{} }
func (s emptyComponentCollection) ToSet() ComponentSet                               { return s }
func (s emptyComponentCollection) Distinct() ComponentCollection                     { return s }
func (s emptyComponentCollection) Contains(Component) bool                           { return false }
func (s emptyComponentCollection) Len() int                                          { return 0 }
func (s emptyComponentCollection) Format(f fmt.State, _ rune)                        { _, _ = fmt.Fprint(f, "[]") }

var _emptyComponents = emptyComponentCollection{}

func EmptyComponents() ComponentCollection {
	return _emptyComponents
}

type componentSet map[Component]struct{}

func newComponentSet(coms ...Component) componentSet {
	cs := componentSet{}
	for _, com := range coms {
		cs.Add(com)
	}

	return cs
}

func (s componentSet) Add(c Component) {
	s[c] = struct{}{}
}

func (s componentSet) Contains(c Component) bool {
	_, ok := s[c]
	return ok
}

func (s componentSet) Len() int {
	return len(s)
}

func (s componentSet) Iterate(f func(Component) bool) bool {
	for c := range s {
		if !f(c) {
			return false
		}
	}
	return true
}

func (s componentSet) Each(f func(Component)) { _componentsEach(s, f) }
func (s componentSet) Filter(predicate func(Component) bool) ComponentCollection {
	return _componentsFilter(s, predicate)
}
func (s componentSet) ToArray() ComponentSlice       { return _componentsToArray(s) }
func (s componentSet) ToSet() ComponentSet           { return s }
func (s componentSet) Distinct() ComponentCollection { return s }
func (s componentSet) Format(fs fmt.State, r rune)   { _componentsFormat(s, fs, r) }

func CombineComponents(its ...ComponentIterator) ComponentCollection {
	var combined ComponentIterator = EmptyComponents()

	for _, it := range its {
		combined = &combinedComponentIterator{
			left:  combined,
			right: it,
		}
	}

	return ComponentsOfIterator(combined)
}
