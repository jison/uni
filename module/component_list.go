package module

type arrayComponentList []Component

func (l arrayComponentList) Each(f func(Component)) {
	for _, c := range l {
		f(c)
	}
}

type emptyComponentList struct{}

func (l emptyComponentList) Each(func(Component)) {}

func emptyComponents() ComponentList {
	return &emptyComponentList{}
}

type filteredComponentList struct {
	componentList ComponentList
	predicate     func(Component) bool
}

func (l *filteredComponentList) Each(f func(Component)) {
	l.componentList.Each(func(c Component) {
		if l.predicate(c) {
			f(c)
		}
	})
}

func filterComponentList(l ComponentList, p func(Component) bool) ComponentList {
	return &filteredComponentList{componentList: l, predicate: p}
}

func ArrayOfComponents(l ComponentList) []Component {
	var arr []Component
	l.Each(func(r Component) {
		arr = append(arr, r)
	})
	return arr
}
