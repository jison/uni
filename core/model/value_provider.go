package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"

	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/location"
	"github.com/jison/uni/internal/reflecting"
)

type valueProvider struct {
	*baseConsumer
	baseProvider
	value reflect.Value
	com   *component
}

var _ Provider = &valueProvider{}

func (vp *valueProvider) Dependencies() DependencyIterator {
	return emptyDependencyIterator{}
}

func (vp *valueProvider) Components() ComponentCollection {
	return ComponentsOfIterator(vp.com)
}

func (vp *valueProvider) Validate() error {
	errs := errors.Empty()

	if !vp.value.IsValid() {
		return errors.Newf("can not provide nil value")
	}

	if !vp.value.CanInterface() {
		errs = errs.AddErrorf("reflect.ValueOf(%v).CanInterface() is false", vp.value)
	}
	if reflecting.IsErrorValue(vp.value) {
		errs = errs.AddErrorf("[%v](%v) is an error, can not provide error value", vp.value.Type(), vp.value)
	}

	if err := vp.com.Validate(); err != nil {
		errs = errs.AddErrors(err)
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func (vp *valueProvider) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "Value[%v](%+v) in %v", vp.value.Type(), vp.value, vp.Scope())

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " at %v", vp.loc)
	}
}

func (vp *valueProvider) clone() *valueProvider {
	cloned := &valueProvider{
		baseConsumer: vp.baseConsumer.clone(),
		baseProvider: vp.baseProvider,
		value:        vp.value,
		com:          vp.com.clone(),
	}
	cloned.com.provider = cloned
	return cloned
}

func (vp *valueProvider) Equal(other interface{}) bool {
	o, ok := other.(*valueProvider)
	if !ok {
		return false
	}
	if vp.baseConsumer != nil {
		if !vp.baseConsumer.Equal(o.baseConsumer) {
			return false
		}
	} else if o.baseConsumer != nil {
		return false
	}

	if vp.value != o.value {
		return false
	}

	if vp.com != nil {
		if !vp.com.Equal(o.com) {
			return false
		}
	} else if o.com != nil {
		return false
	}

	return true
}

func (vp *valueProvider) String() string {
	return fmt.Sprintf("%v", vp)
}

type ValueProviderBuilder interface {
	ModuleOption
	ProviderBuilder

	SetIgnore(ignore bool) ValueProviderBuilder
	SetHidden(hidden bool) ValueProviderBuilder
	AddAs(ifs ...TypeVal) ValueProviderBuilder
	SetName(name string) ValueProviderBuilder
	AddTags(tags ...Symbol) ValueProviderBuilder

	SetScope(scope Scope) ValueProviderBuilder
	SetLocation(loc location.Location) ValueProviderBuilder
	UpdateCallLocation(loc location.Location) ValueProviderBuilder
}

func (vp *valueProvider) ApplyModule(mb ModuleBuilder) {
	mb.AddProvider(vp)
}

func (vp *valueProvider) Provider() Provider {
	return vp.clone()
}

func (vp *valueProvider) SetIgnore(ignore bool) ValueProviderBuilder {
	vp.com.SetIgnore(ignore)
	return vp
}

func (vp *valueProvider) SetHidden(hidden bool) ValueProviderBuilder {
	vp.com.SetHidden(hidden)
	return vp
}
func (vp *valueProvider) AddAs(ifs ...TypeVal) ValueProviderBuilder {
	vp.com.AddAs(ifs...)
	return vp
}
func (vp *valueProvider) SetName(name string) ValueProviderBuilder {
	vp.com.SetName(name)
	return vp
}
func (vp *valueProvider) AddTags(tags ...Symbol) ValueProviderBuilder {
	vp.com.AddTags(tags...)
	return vp
}

func (vp *valueProvider) SetScope(scope Scope) ValueProviderBuilder {
	vp.baseConsumer.SetScope(scope)
	return vp
}

func (vp *valueProvider) SetLocation(loc location.Location) ValueProviderBuilder {
	vp.loc = loc
	return vp
}

func (vp *valueProvider) UpdateCallLocation(loc location.Location) ValueProviderBuilder {
	if vp.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		vp.SetLocation(loc)
	}
	return vp
}

func valueProviderOf(val interface{}, opts ...ValueProviderOption) *valueProvider {
	var rVal reflect.Value
	var rType reflect.Type
	if val == nil {
		rVal = reflect.Value{}
	} else {
		rVal = reflect.ValueOf(val)
		rType = rVal.Type()
	}

	vp := &valueProvider{
		baseConsumer: &baseConsumer{
			val: valuer.Const(rVal),
		},
		value: rVal,
		com: &component{
			val:   valuer.Identity(),
			rType: rType,
		},
	}
	vp.com.provider = vp

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyValueProvider(vp)
	}
	return vp
}

func Value(val interface{}, opts ...ValueProviderOption) ValueProviderBuilder {
	vp := valueProviderOf(val, opts...).UpdateCallLocation(nil)

	return vp
}

type ValueProviderOption interface {
	ApplyValueProvider(b ValueProviderBuilder)
}

func (o LocationOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.SetLocation(o.Location)
}

func (o UpdateCallLocationOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.UpdateCallLocation(o.Location)
}

func (o ScopeOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.SetScope(o.scope)
}

func (o IgnoreOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.SetIgnore(o.ignore)
}

func (o HiddenOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.SetHidden(o.hidden)
}

func (o AsOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.AddAs(o.ifs...)
}

func (o NameOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.SetName(string(o))
}

func (o TagsOption) ApplyValueProvider(b ValueProviderBuilder) {
	b.AddTags(o.tags...)
}
