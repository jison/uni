package model

import (
	"reflect"
)

type ComponentRepository interface {
	AllComponents() ComponentCollection
	ComponentsMatch(Criteria) ComponentCollection
	ComponentsWithScope(scope Scope) ComponentCollection
	ComponentsMatchDependency(dep Dependency) ComponentCollection
}

type componentRepository struct {
	entryByType       map[reflect.Type]typeEntry
	componentsByScope map[Scope]componentSet
	allComponents     componentSet
}

var _ ComponentRepository = &componentRepository{}

type typeEntry struct {
	componentsByName  map[string]componentSet
	componentsByTag   map[Symbol]componentSet
	exposedComponents componentSet
}

func (m *componentRepository) AllComponents() ComponentCollection {
	return m.allComponents
}

func (m *componentRepository) ComponentsMatch(cri Criteria) ComponentCollection {
	if cri == nil {
		return EmptyComponents()
	}
	if isCriteriaMatchAll(cri) {
		return m.allComponents
	}

	if entry, ok := m.entryByType[cri.Type()]; ok {
		var components ComponentCollection
		if cri.Name() != "" {
			if components, ok = entry.componentsByName[cri.Name()]; ok {
				return components.Filter(func(com Component) bool { return componentMatch(com, cri) })
			}
		} else if cri.Tags().Len() > 0 {
			var firstTag Symbol
			cri.Tags().Iterate(func(t Symbol) bool {
				firstTag = t
				return false
			})

			if components, ok = entry.componentsByTag[firstTag]; ok {
				return components.Filter(func(com Component) bool { return componentMatch(com, cri) })
			}
		} else {
			return entry.exposedComponents
		}
	}

	return EmptyComponents()
}

func (m *componentRepository) ComponentsWithScope(scope Scope) ComponentCollection {
	if scope == nil {
		scope = GlobalScope
	}

	if set, ok := m.componentsByScope[scope]; ok {
		return set
	} else {
		return EmptyComponents()
	}
}

func (m *componentRepository) ComponentsMatchDependency(dep Dependency) ComponentCollection {
	if dep == nil {
		return EmptyComponents()
	}

	s := dep.Consumer().Scope()

	if isCriteriaMatchAll(dep) {
		return m.ComponentsWithScope(s)
	}

	return m.ComponentsMatch(dep).Filter(func(com Component) bool {
		// can not use the same provider as dependence's consumer
		if com.Provider() == dep.Consumer() {
			return false
		}

		comScope := com.Provider().Scope()

		// the scope with the injected component could enter the dependence's scope
		if s != comScope && !s.CanEnterFrom(comScope) {
			return false
		}

		return true
	})
}

func (m *componentRepository) typeEntryOf(t reflect.Type) typeEntry {
	var entry typeEntry
	var ok bool
	if entry, ok = m.entryByType[t]; !ok {
		entry = typeEntry{
			componentsByName:  map[string]componentSet{},
			componentsByTag:   map[Symbol]componentSet{},
			exposedComponents: newComponentSet(),
		}
		m.entryByType[t] = entry
	}
	return entry
}

func (m *componentRepository) addComponent(com Component) {
	if com.Ignored() {
		return
	}
	if m.allComponents.Contains(com) {
		return
	}
	m.allComponents.Add(com)

	m.addComponentInScope(com, com.Provider().Scope())
	m.addComponentAsType(com, com.Type())

	com.As().Iterate(func(t reflect.Type) bool {
		m.addComponentAsType(com, t)
		return true
	})
}

func (m *componentRepository) addComponentInScope(com Component, scope Scope) {
	var set componentSet
	var ok bool
	if set, ok = m.componentsByScope[scope]; !ok {
		set = componentSet{}
		m.componentsByScope[scope] = set
	}

	set.Add(com)
}

func (m *componentRepository) addComponentAsType(com Component, t reflect.Type) {
	entry := m.typeEntryOf(t)

	var comSet componentSet
	var ok bool

	if com.Tags().Len() > 0 {
		com.Tags().Iterate(func(tag Symbol) bool {
			if comSet, ok = entry.componentsByTag[tag]; !ok {
				comSet = newComponentSet()
				entry.componentsByTag[tag] = comSet
			}
			comSet.Add(com)

			return true
		})
	}

	if com.Name() != "" {
		if comSet, ok = entry.componentsByName[com.Name()]; !ok {
			comSet = newComponentSet()
			entry.componentsByName[com.Name()] = comSet
		}

		comSet.Add(com)
	}

	if !com.Hidden() {
		entry.exposedComponents.Add(com)
	}
}

func componentMatch(com Component, cri Criteria) bool {
	if com == nil || cri == nil {
		return false
	}

	if !com.As().Has(cri.Type()) && com.Type() != cri.Type() {
		return false
	}

	if cri.Name() != "" {
		if com.Name() != cri.Name() {
			return false
		}
	}

	if cri.Tags().Len() > 0 {
		match := true
		cri.Tags().Iterate(func(s Symbol) bool {
			if !com.Tags().Has(s) {
				match = false
				return false
			}
			return true
		})

		if !match {
			return false
		}
	}

	if com.Hidden() && cri.Name() == "" && cri.Tags().Len() == 0 {
		return false
	}

	return true
}

func isCriteriaMatchAll(cri Criteria) bool {
	return IsWildCardType(cri.Type()) && cri.Name() == "" && cri.Tags().Len() == 0
}

func NewRepository(components ComponentCollection) ComponentRepository {
	matcher := &componentRepository{
		entryByType:       map[reflect.Type]typeEntry{},
		componentsByScope: map[Scope]componentSet{},
		allComponents:     componentSet{},
	}

	components.Each(func(com Component) {
		matcher.addComponent(com)
	})

	return matcher
}
