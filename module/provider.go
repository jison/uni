package module

import (
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type componentMap map[int]*component

func (cl componentMap) Each(f func(Component)) {
	for _, c := range cl {
		f(c)
	}
}

type location struct {
	file string
	line int
}

func (l location) File() string {
	return l.file
}

func (l location) Line() int {
	return l.line
}

type provider struct {
	funcVal           reflect.Value
	location          location
	scope             string
	parameterList     parameterList
	componentList     componentMap
	fakeParameterList map[int]*parameter
	fakeComponentList map[int]*component
}

var _ Provider = &provider{}

func (p *provider) FuncVal() reflect.Value {
	return p.funcVal
}

func (p *provider) Scope() string {
	return p.scope
}

func (p *provider) Location() Location {
	return p.location
}

func (p *provider) Parameters() ParameterList {
	return p.parameterList
}

func (p *provider) Components() ComponentList {
	return p.componentList
}

func (p *provider) Validate() error {
	errs := make([]error, 0)

	// TODO: funcVal is not a funcVal

	for _, param := range p.parameterList {
		err := param.Validate()
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, fakeParam := range p.fakeParameterList {
		err := errors.New("invalid parameter index :%d", fakeParam.index)
		errs = append(errs, err)
	}
	for _, com := range p.componentList {
		err := com.Validate()
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, fakeCom := range p.fakeComponentList {
		err := errors.New("invalid component index :%d", fakeCom.index)
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Merge(errs...)
}

type providerOption interface {
	applyProvider(*provider)
}

type scopeOption string

func (o scopeOption) applyProvider(p *provider) {
	p.scope = string(o)
}

func Scope(name string) scopeOption {
	return scopeOption(name)
}

type paramOption struct {
	index     int
	paramOpts []parameterOption
}

func (o *paramOption) applyProvider(p *provider) {
	var param *parameter
	if o.index < len(p.parameterList) {
		param = p.parameterList[o.index]
	} else {
		if fp, ok := p.fakeParameterList[o.index]; ok {
			param = fp
		} else {
			param = &parameter{
				index: o.index,
				rtype: reflect.TypeOf(nil),
			}
			p.fakeParameterList[o.index] = param
		}
	}

	for _, po := range o.paramOpts {
		po.applyParameter(param)
	}
}

func Param(index int, paramOpts ...parameterOption) *paramOption {
	return &paramOption{index, paramOpts}
}

type comOption struct {
	index   int
	comOpts []componentOption
}

func (o *comOption) applyProvider(p *provider) {
	var com *component
	if c, ok := p.componentList[o.index]; ok {
		com = c
	} else {
		if fc, ok := p.fakeComponentList[o.index]; ok {
			com = fc
		} else {
			com := &component{
				provider: p,
				index:    o.index,
				rtype:    reflect.TypeOf(nil),
			}
			p.fakeComponentList[o.index] = com
		}
	}

	for _, co := range o.comOpts {
		co.applyComponent(com)
	}
}

func Comp(index int, comOpts ...componentOption) *comOption {
	return &comOption{index, comOpts}
}

func (o location) applyProvider(p *provider) {
	p.location = o
}
