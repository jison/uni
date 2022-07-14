package module

type componentSet map[Component]struct{}

func newComponentSet() componentSet {
	return make(componentSet, 0)
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

func (s componentSet) Head() Component {
	for c := range s {
		return c
	}
	return nil
}

func (s componentSet) Each(f func(Component)) {
	for c := range s {
		f(c)
	}
}
