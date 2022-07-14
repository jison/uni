package module

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type matcher struct {
	entryByType map[reflect.Type]typeEntry
	components  componentSet
}

type typeEntry struct {
	componentsByName  map[string]componentSet
	componentsByTag   map[Symbol]componentSet
	exposedComponents componentSet
}

var _ ComponentMatcher = &matcher{}

func (m *matcher) Match(c Criteria, scope string) ComponentList {
	if c == nil {
		return emptyComponents()
	}

	// TODO: handle scope
	if entry, ok := m.entryByType[c.Type()]; ok {
		if c.Name() != "" {
			if coms, ok := entry.componentsByName[c.Name()]; ok {
				return filterComponentList(coms, func(com Component) bool { return com.Match(c) })
			}
		} else if len(c.Tags()) > 0 {
			if coms, ok := entry.componentsByTag[c.Tags()[0]]; ok {
				return filterComponentList(coms, func(com Component) bool { return com.Match(c) })
			}
		} else {
			return entry.exposedComponents
		}
	}

	return emptyComponents()
}

func (m *matcher) All() ComponentList {
	return m.components
}

type dupNameError struct {
	rtype      reflect.Type
	name       string
	components []Component
}

func (e *dupNameError) Error() string {
	return fmt.Sprintf("duplicate name `%s` of type `%v`", e.name, e.rtype)
}

func MatcherOfComponents(cl ComponentList) (ComponentMatcher, error) {
	duplicateNames := make(map[reflect.Type]string)
	entryByType := make(map[reflect.Type]typeEntry, 0)
	allComponents := newComponentSet()

	cl.Each(func(com Component) {
		if allComponents.Contains(com) {
			return
		}
		allComponents.Add(com)

		for _, t := range com.AllTypes() {
			var entry typeEntry
			var ok bool
			if entry, ok = entryByType[t]; !ok {
				entry = typeEntry{
					componentsByName:  make(map[string]componentSet, 0),
					componentsByTag:   make(map[Symbol]componentSet, 0),
					exposedComponents: newComponentSet(),
				}
				entryByType[t] = entry
			}

			if len(com.Tags()) > 0 {
				for _, tag := range com.Tags() {
					var coms componentSet
					if coms, ok = entry.componentsByTag[tag]; !ok {
						coms = newComponentSet()
						entry.componentsByTag[tag] = coms
					}
					coms.Add(com)
				}
			}

			if com.Name() != "" {
				if coms, ok := entry.componentsByName[com.Name()]; !ok {
					coms = newComponentSet()
					coms.Add(com)
				} else {
					coms.Add(com)
					if coms.Len() > 1 {
						duplicateNames[t] = com.Name()
					}
				}
			}

			if !com.Hidden() {
				entry.exposedComponents.Add(com)
			}
		}
	})

	if len(duplicateNames) > 0 {
		errs := make([]error, 0, len(duplicateNames))
		for rtype, name := range duplicateNames {
			coms := entryByType[rtype].componentsByName[name]
			err := &dupNameError{rtype, name, ArrayOfComponents(coms)}
			errs = append(errs, err)
		}
		return nil, errors.Merge(errs...)
	}

	return &matcher{entryByType: entryByType, components: allComponents}, nil
}
