package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"

	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/reflecting"
)

type componentByIndex map[int]*component

func (m componentByIndex) Iterate(f func(Component) bool) bool {
	for _, c := range m {
		if !f(c) {
			return false
		}
	}
	return true
}

type funcProvider struct {
	*funcConsumer
	baseProvider

	components     componentByIndex
	fakeComponents componentByIndex
}

var _ Provider = &funcProvider{}

func (fp *funcProvider) Components() ComponentCollection {
	return ComponentsOfIterator(fp.components)
}

func (fp *funcProvider) Validate() error {
	errs := errors.Empty()

	if !fp.funcVal.IsValid() {
		return errors.Newf("function is nil")
	}

	if err := fp.funcConsumer.Validate(); err != nil {
		var structErr errors.StructError
		if errors.As(err, &structErr) {
			errs = structErr
		}
		// it must be StructError
		//else {
		//	errs = errs.AddErrors(err)
		//}
	}

	for index := range fp.fakeComponents {
		errs = errs.AddErrorf("[%v] return value at index %v is nonexistent or is not a valid type",
			fp.funcVal.Type(), index)
	}
	if len(fp.components) == 0 {
		errs = errs.AddErrorf("[%v] does not return any valid value", fp.funcVal.Type())
	}
	for index, com := range fp.components {
		if err := com.Validate(); err != nil {
			var structErr errors.StructError
			if errors.As(err, &structErr) {
				err = structErr.WithMainf("return value at %v", index)
			}
			errs = errs.AddErrors(err)
		}
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func (fp *funcProvider) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "Function[%v] in %v", fp.funcVal.Type(), fp.Scope())

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " at %v", fp.loc)
	}
}

func (fp *funcProvider) clone() *funcProvider {
	if fp == nil {
		return nil
	}

	newFP := &funcProvider{
		funcConsumer:   fp.funcConsumer.clone(),
		baseProvider:   fp.baseProvider,
		components:     componentByIndex{},
		fakeComponents: componentByIndex{},
	}

	for _, param := range newFP.funcConsumer.params {
		param.consumer = newFP
	}
	for _, param := range newFP.funcConsumer.fakeParams {
		param.consumer = newFP
	}

	cloneComponents := func(oldComponents, newComponents componentByIndex) {
		for i, com := range oldComponents {
			newCom := com.clone()
			newCom.provider = newFP
			newComponents[i] = newCom
		}
	}

	cloneComponents(fp.components, newFP.components)
	cloneComponents(fp.fakeComponents, newFP.fakeComponents)

	return newFP
}

func (fp *funcProvider) Equal(other interface{}) bool {
	o, ok := other.(*funcProvider)
	if !ok {
		return false
	}
	if fp == nil || o == nil {
		return fp == nil && o == nil
	}

	if fp.funcConsumer != nil {
		if !fp.funcConsumer.Equal(o.funcConsumer) {
			return false
		}
	} else if o.funcConsumer != nil {
		return false
	}

	if len(fp.components) != len(o.components) {
		return false
	}
	for i, com := range fp.components {
		if com2, comOk := o.components[i]; !comOk {
			return false
		} else if !com.Equal(com2) {
			return false
		}
	}

	if len(fp.fakeComponents) != len(o.fakeComponents) {
		return false
	}
	for i, com := range fp.fakeComponents {
		if com2, comOk := o.fakeComponents[i]; !comOk {
			return false
		} else if !com.Equal(com2) {
			return false
		}
	}

	return true
}

type FuncProviderBuilder interface {
	ModuleOption
	ProviderBuilder
	Param(index int, opts ...DependencyOption) FuncProviderBuilder
	Return(index int, opts ...ComponentOption) FuncProviderBuilder
	SetScope(scope Scope) FuncProviderBuilder
	SetLocation(loc location.Location) FuncProviderBuilder
	UpdateCallLocation(loc location.Location) FuncProviderBuilder
}

func (fp *funcProvider) ApplyModule(b ModuleBuilder) {
	b.AddProvider(fp)
}

func (fp *funcProvider) Provider() Provider {
	return fp.clone()
}

func (fp *funcProvider) Param(index int, opts ...DependencyOption) FuncProviderBuilder {
	fp.funcConsumer.Param(index, opts...)
	return fp
}

func (fp *funcProvider) Return(index int, opts ...ComponentOption) FuncProviderBuilder {
	var com *component
	var ok bool
	if com, ok = fp.components[index]; !ok {
		if com, ok = fp.fakeComponents[index]; !ok {
			com = &component{
				provider: fp,
				rType:    reflect.TypeOf(nil),
			}
			fp.fakeComponents[index] = com
		}
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyComponent(com)
	}

	return fp
}

func (fp *funcProvider) SetScope(scope Scope) FuncProviderBuilder {
	fp.baseConsumer.SetScope(scope)
	return fp
}

func (fp *funcProvider) SetLocation(loc location.Location) FuncProviderBuilder {
	fp.funcConsumer.SetLocation(loc)
	return fp
}

func (fp *funcProvider) UpdateCallLocation(loc location.Location) FuncProviderBuilder {
	if fp.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		fp.SetLocation(loc)
	}
	return fp
}

func funcProviderOf(val interface{}, opts ...FuncProviderOption) *funcProvider {
	var funcType reflect.Type
	if val != nil {
		funcType = reflect.TypeOf(val)
	}

	p := &funcProvider{
		funcConsumer:   funcConsumerOf(val),
		baseProvider:   baseProvider{},
		components:     componentByIndex{},
		fakeComponents: componentByIndex{},
	}

	for _, param := range p.funcConsumer.params {
		param.consumer = p
	}

	if funcType != nil && funcType.Kind() == reflect.Func {
		for i := 0; i < funcType.NumOut(); i++ {
			comType := funcType.Out(i)
			if reflecting.IsErrorType(comType) {
				continue
			}
			p.components[i] = &component{
				provider: p,
				val:      valuer.Index(i),
				rType:    comType,
			}
		}
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyFuncProvider(p)
	}

	return p
}

func Func(function interface{}, opts ...FuncProviderOption) FuncProviderBuilder {
	fp := funcProviderOf(function, opts...).UpdateCallLocation(nil)
	return fp
}

type FuncProviderOption interface {
	ApplyFuncProvider(FuncProviderBuilder)
}

func (o ParamOption) ApplyFuncProvider(b FuncProviderBuilder) {
	b.Param(o.index, o.opts...)
}

func (o ReturnOption) ApplyFuncProvider(b FuncProviderBuilder) {
	b.Return(o.index, o.opts...)
}

func (o LocationOption) ApplyFuncProvider(b FuncProviderBuilder) {
	b.SetLocation(o.Location)
}

func (o UpdateCallLocationOption) ApplyFuncProvider(b FuncProviderBuilder) {
	b.UpdateCallLocation(o.Location)
}

func (o ScopeOption) ApplyFuncProvider(b FuncProviderBuilder) {
	b.SetScope(o.scope)
}
