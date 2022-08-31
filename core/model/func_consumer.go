package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/location"
)

type funcParam struct {
	*dependency
	index int
}

var _ Dependency = &funcParam{}

func (p *funcParam) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "%+v at parameter `%d`", p.dependency, p.index)

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " of %+v", p.Consumer())
	}
}

func (p *funcParam) clone() *funcParam {
	return &funcParam{
		dependency: p.dependency.clone(),
		index:      p.index,
	}
}

func (p *funcParam) Equal(other interface{}) bool {
	o, ok := other.(*funcParam)
	if !ok {
		return false
	}
	if p == nil || o == nil {
		return p == nil && o == nil
	}
	if p.index != o.index {
		return false
	}
	if p.dependency != nil {
		if !p.dependency.Equal(o.dependency) {
			return false
		}
	} else if o.dependency != nil {
		return false
	}

	return true
}

type paramByIndex map[int]*funcParam

func (m paramByIndex) Iterate(f func(Dependency) bool) bool {
	for _, d := range m {
		if !f(d) {
			return false
		}
	}
	return true
}

type funcConsumer struct {
	*baseConsumer
	funcVal    reflect.Value
	params     paramByIndex
	fakeParams paramByIndex
}

var _ Consumer = &funcConsumer{}

func (fc *funcConsumer) Dependencies() DependencyIterator {
	return fc.params
}

func (fc *funcConsumer) Validate() error {
	errs := errors.Empty()

	if !fc.funcVal.IsValid() {
		return errors.Newf("function is nil")
	}

	if fc.funcVal.Kind() != reflect.Func {
		errs = errs.AddErrorf("[%v](%v) is not a function", fc.funcVal.Type(), fc.funcVal)
	}

	for _, fakeParam := range fc.fakeParams {
		errs = errs.AddErrorf("param at index `%v` is nonexistent", fakeParam.index)
	}
	for _, param := range fc.params {
		if err := param.Validate(); err != nil {
			var structErr errors.StructError
			if errors.As(err, &structErr) {
				err = structErr.WithMainf("param %v", param.index)
			}

			errs = errs.AddErrors(err)
		}
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func (fc *funcConsumer) Format(f fmt.State, r rune) {
	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, "FunctionConsumer[%v] at %v", fc.funcVal.Type(), fc.Location())
	} else {
		_, _ = fmt.Fprintf(f, "FunctionConsumer[%v]", fc.funcVal.Type())
	}
}

func (fc *funcConsumer) clone() *funcConsumer {
	cloned := &funcConsumer{
		baseConsumer: fc.baseConsumer.clone(),
		funcVal:      fc.funcVal,
		params:       paramByIndex{},
		fakeParams:   paramByIndex{},
	}

	cloneParams := func(oldParams, newParams paramByIndex) {
		for i, param := range oldParams {
			p := param.clone()
			p.consumer = cloned

			newParams[i] = p
		}
	}
	cloneParams(fc.params, cloned.params)
	cloneParams(fc.fakeParams, cloned.fakeParams)

	return cloned
}

func (fc *funcConsumer) Equal(other interface{}) bool {
	o, ok := other.(*funcConsumer)
	if !ok {
		return false
	}
	if fc.funcVal != o.funcVal {
		return false
	}
	if len(fc.params) != len(o.params) {
		return false
	}
	for index, p := range fc.params {
		if p2, pok := o.params[index]; !pok {
			return false
		} else if !p2.Equal(p) {
			return false
		}
	}
	if len(fc.fakeParams) != len(o.fakeParams) {
		return false
	}
	for index, p := range fc.fakeParams {
		if p2, pok := o.fakeParams[index]; !pok {
			return false
		} else if !p2.Equal(p) {
			return false
		}
	}

	if fc.baseConsumer != nil {
		if !fc.baseConsumer.Equal(o.baseConsumer) {
			return false
		}
	} else if o.baseConsumer != nil {
		return false
	}

	return true
}

type FuncConsumerBuilder interface {
	Param(index int, opts ...DependencyOption) FuncConsumerBuilder
	SetScope(scope Scope) FuncConsumerBuilder
	SetLocation(loc location.Location) FuncConsumerBuilder
	UpdateCallLocation(loc location.Location) FuncConsumerBuilder
	Consumer() Consumer
}

func (fc *funcConsumer) Param(index int, opts ...DependencyOption) FuncConsumerBuilder {
	var param *funcParam
	var ok bool
	if param, ok = fc.params[index]; !ok {
		if param, ok = fc.fakeParams[index]; !ok {
			param = &funcParam{
				dependency: &dependency{
					consumer: fc,
					val:      valuer.Param(index),
					rType:    reflect.TypeOf(nil),
				},
				index: index,
			}
			fc.fakeParams[index] = param
		}
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyDependency(param)
	}

	return fc
}

func (fc *funcConsumer) SetScope(scope Scope) FuncConsumerBuilder {
	fc.baseConsumer.SetScope(scope)
	return fc
}

func (fc *funcConsumer) SetLocation(loc location.Location) FuncConsumerBuilder {
	fc.baseConsumer.SetLocation(loc)
	return fc
}

func (fc *funcConsumer) UpdateCallLocation(loc location.Location) FuncConsumerBuilder {
	if fc.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		fc.SetLocation(loc)
	}
	return fc
}

func (fc *funcConsumer) Consumer() Consumer {
	return fc.clone()
}

func FuncConsumer(val interface{}, opts ...FuncConsumerOption) FuncConsumerBuilder {
	fc := funcConsumerOf(val, opts...).UpdateCallLocation(nil)
	return fc
}

func funcConsumerOf(val interface{}, opts ...FuncConsumerOption) *funcConsumer {
	var funcVal reflect.Value
	var funcType reflect.Type
	if val != nil {
		funcVal = reflect.ValueOf(val)
		funcType = funcVal.Type()
	} else {
		funcVal = reflect.Value{}
	}

	c := &funcConsumer{
		baseConsumer: &baseConsumer{
			val: valuer.Func(funcVal),
		},
		funcVal:    funcVal,
		params:     paramByIndex{},
		fakeParams: paramByIndex{},
	}

	if funcType != nil && funcType.Kind() == reflect.Func {
		for i := 0; i < funcType.NumIn(); i++ {
			paramType := funcType.In(i)
			fParam := &funcParam{
				dependency: &dependency{
					consumer: c,
					val:      valuer.Param(i),
					rType:    paramType,
				},
				index: i,
			}

			if i == funcType.NumIn()-1 && funcType.IsVariadic() {
				// variadic function as collector by default
				fParam.isCollector = true
			}

			c.params[i] = fParam
		}
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyFuncConsumer(c)
	}

	return c
}

type FuncConsumerOption interface {
	ApplyFuncConsumer(builder FuncConsumerBuilder)
}

func (o ParamOption) ApplyFuncConsumer(b FuncConsumerBuilder) {
	b.Param(o.index, o.opts...)
}

func (o ScopeOption) ApplyFuncConsumer(b FuncConsumerBuilder) {
	b.SetScope(o.scope)
}

func (o LocationOption) ApplyFuncConsumer(b FuncConsumerBuilder) {
	b.SetLocation(o.Location)
}

func (o UpdateCallLocationOption) ApplyFuncConsumer(b FuncConsumerBuilder) {
	b.UpdateCallLocation(o.Location)
}
