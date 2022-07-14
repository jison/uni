package module

import (
	"reflect"
)

type Location interface {
	File() string
	Line() int
}

type Provider interface {
	FuncVal() reflect.Value
	Scope() string
	Parameters() ParameterList
	Components() ComponentList
	Location() Location
	Validate() error
}

type ParameterList interface {
	Each(f func(p Parameter))
}

type Criteria interface {
	Type() reflect.Type
	Name() string
	Tags() []Symbol
}

type Parameter interface {
	Criteria
	Provider() Provider
	Optional() bool
	IsCollector() bool
	Index() int
	Validate() error
}

type ComponentList interface {
	Each(f func(c Component))
}

type ComponentMatcher interface {
	Match(criteria Criteria, scope string) ComponentList
	All() ComponentList
}

type Component interface {
	Provider() Provider
	Ignored() bool
	Hidden() bool
	Index() int
	Type() reflect.Type
	As() []reflect.Type
	AllTypes() []reflect.Type
	Name() string
	Tags() []Symbol
	HasTag(Symbol) bool
	Match(Criteria) bool
	Validate() error
}

type Module interface {
	Components() ComponentList
}
