package module

import (
	"reflect"
)

type parameter struct {
	provider    *provider
	optional    bool
	isCollector bool
	index       int
	rtype       reflect.Type
	name        string
	tags        []Symbol
}

var _ Parameter = &parameter{}

func (p *parameter) Provider() Provider {
	return p.provider
}

func (p *parameter) Optional() bool {
	return p.optional
}

func (p *parameter) IsCollector() bool {
	return p.isCollector
}

func (p *parameter) Index() int {
	return p.index
}

func (p *parameter) Type() reflect.Type {
	return p.rtype
}

func (p *parameter) Name() string {
	return p.name
}

func (p *parameter) Tags() []Symbol {
	return p.tags
}

func (p *parameter) Validate() error {
	// TODO: validate
	// can not add error type parameter
	// collector only apply for slice or array type

	return nil
}

type parameterList []*parameter

var _ ParameterList = parameterList{}

func (l parameterList) Each(f func(p Parameter)) {
	for _, p := range l {
		f(p)
	}
}

type parameterOption interface {
	applyParameter(*parameter)
}

type optionalOption bool

func (o optionalOption) applyParameter(p *parameter) {
	p.optional = bool(o)
}

func Optional(b bool) parameterOption {
	return optionalOption(b)
}

type asCollectorOption struct{}

func (o asCollectorOption) applyParameter(p *parameter) {
	p.isCollector = true
}

func AsCollector() parameterOption {
	return asCollectorOption{}
}
