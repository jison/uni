package model

import (
	"reflect"

	"github.com/jison/uni/internal/location"
)

func Name(n string) NameOption {
	return NameOption(n)
}

type NameOption string

func ByName(n string) ByNameOption {
	return ByNameOption(n)
}

type ByNameOption string

func Tags(tags ...Symbol) TagsOption {
	return TagsOption{tags: tags}
}

type TagsOption struct {
	tags []Symbol
}

func ByTags(tags ...Symbol) ByTagsOption {
	return ByTagsOption{tags: tags}
}

type ByTagsOption struct {
	tags []Symbol
}

func Ignore() IgnoreOption {
	return IgnoreOption{ignore: true}
}

type IgnoreOption struct {
	ignore bool
}

func Hide() HiddenOption {
	return HiddenOption{hidden: true}
}

type HiddenOption struct {
	hidden bool
}

func As(ifs ...TypeVal) AsOption {
	return AsOption{ifs}
}

type AsOption struct {
	ifs []TypeVal
}

func Optional(optional bool) OptionalOption {
	return OptionalOption(optional)
}

type OptionalOption bool

func AsCollector(as bool) AsCollectorOption {
	return AsCollectorOption(as)
}

type AsCollectorOption bool

func Location(loc location.Location) LocationOption {
	return LocationOption{loc}
}

type LocationOption struct {
	location.Location
}

func UpdateCallLocation() UpdateCallLocationOption {
	loc := location.GetCallLocation(3).Callee()
	return UpdateCallLocationOption{loc}
}

type UpdateCallLocationOption struct {
	location.Location
}

func InScope(scope Scope) ScopeOption {
	return ScopeOption{scope}
}

type ScopeOption struct {
	scope Scope
}

func Field(name string, opts ...DependencyOption) FieldOption {
	return FieldOption{
		name: name,
		opts: opts,
	}
}

type FieldOption struct {
	name string
	opts []DependencyOption
}

func IgnoreFields(p func(field reflect.StructField) bool) IgnoreFieldsOption {
	return IgnoreFieldsOption{p}
}

type IgnoreFieldsOption struct {
	predicate func(field reflect.StructField) bool
}

func Param(index int, opts ...DependencyOption) ParamOption {
	return ParamOption{index: index, opts: opts}
}

type ParamOption struct {
	index int
	opts  []DependencyOption
}

func Return(index int, opts ...ComponentOption) ReturnOption {
	return ReturnOption{
		index: index,
		opts:  opts,
	}
}

type ReturnOption struct {
	index int
	opts  []ComponentOption
}

type WithScopeOption struct {
	scope Scope
	pbs   []ProviderBuilder
}

func WithScope(scope Scope) func(...ProviderBuilder) *WithScopeOption {
	return func(pbs ...ProviderBuilder) *WithScopeOption {
		return &WithScopeOption{
			scope: scope,
			pbs:   pbs,
		}
	}
}

func (o *WithScopeOption) ApplyModule(mb ModuleBuilder) {
	for _, pb := range o.pbs {
		switch pp := pb.(type) {
		case FuncProviderBuilder:
			pp.SetScope(o.scope)
		case StructProviderBuilder:
			pp.SetScope(o.scope)
		case ValueProviderBuilder:
			pp.SetScope(o.scope)
		}

		mb.AddProvider(pb)
	}
}
