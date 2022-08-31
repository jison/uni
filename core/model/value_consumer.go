package model

import (
	"fmt"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"
)

type valueConsumer struct {
	*dependency
	*baseConsumer
}

func (v *valueConsumer) Dependencies() DependencyIterator {
	return v.dependency
}

func (v *valueConsumer) Valuer() valuer.Valuer {
	return v.baseConsumer.val
}

func (v *valueConsumer) Validate() error {
	return v.dependency.Validate()
}

func (v *valueConsumer) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "ValueConsumer[%v]", v.dependency.Type())

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " at %v", v.Location())
	}
}

func (v *valueConsumer) clone() *valueConsumer {
	cloned := &valueConsumer{
		dependency:   v.dependency.clone(),
		baseConsumer: v.baseConsumer.clone(),
	}
	cloned.dependency.consumer = cloned

	return cloned
}

func (v *valueConsumer) Equal(other interface{}) bool {
	o, ok := other.(*valueConsumer)
	if !ok {
		return false
	}
	if v == nil || o == nil {
		return v == nil && o == nil
	}

	if v.dependency != nil {
		if !v.dependency.Equal(o.dependency) {
			return false
		}
	} else if o.dependency != nil {
		return false
	}

	if v.baseConsumer != nil {
		if !v.baseConsumer.Equal(o.baseConsumer) {
			return false
		}
	} else if o.baseConsumer != nil {
		return false
	}

	return true
}

type ValueConsumerBuilder interface {
	SetOptional(optional bool) ValueConsumerBuilder
	SetAsCollector(asCollector bool) ValueConsumerBuilder
	SetName(name string) ValueConsumerBuilder
	AddTags(tags ...Symbol) ValueConsumerBuilder
	SetScope(scope Scope) ValueConsumerBuilder
	SetLocation(loc location.Location) ValueConsumerBuilder
	UpdateCallLocation(loc location.Location) ValueConsumerBuilder
	Consumer() Consumer
}

func (v *valueConsumer) SetOptional(optional bool) ValueConsumerBuilder {
	v.dependency.SetOptional(optional)
	return v
}

func (v *valueConsumer) SetAsCollector(asCollector bool) ValueConsumerBuilder {
	v.dependency.SetAsCollector(asCollector)
	return v
}

func (v *valueConsumer) SetName(name string) ValueConsumerBuilder {
	v.dependency.SetName(name)
	return v
}

func (v *valueConsumer) AddTags(tags ...Symbol) ValueConsumerBuilder {
	v.dependency.AddTags(tags...)
	return v
}

func (v *valueConsumer) SetScope(scope Scope) ValueConsumerBuilder {
	v.baseConsumer.SetScope(scope)
	return v
}

func (v *valueConsumer) SetLocation(loc location.Location) ValueConsumerBuilder {
	v.baseConsumer.SetLocation(loc)
	return v
}

func (v *valueConsumer) UpdateCallLocation(loc location.Location) ValueConsumerBuilder {
	if v.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		v.SetLocation(loc)
	}
	return v
}

func (v *valueConsumer) Consumer() Consumer {
	return v.clone()
}

func valueConsumerOf(t TypeVal, opts ...ValueConsumerOption) *valueConsumer {
	vc := &valueConsumer{
		baseConsumer: &baseConsumer{
			val: valuer.Identity(),
		},
		dependency: &dependency{
			rType: TypeOf(t),
			val:   valuer.Identity(),
		},
	}
	vc.dependency.consumer = vc

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyValueConsumer(vc)
	}

	return vc
}

func ValueConsumer(t TypeVal, opts ...ValueConsumerOption) ValueConsumerBuilder {
	vc := valueConsumerOf(t, opts...).UpdateCallLocation(nil)
	return vc
}

type ValueConsumerOption interface {
	ApplyValueConsumer(b ValueConsumerBuilder)
}

func (o OptionalOption) ApplyValueConsumer(b ValueConsumerBuilder) {
	b.SetOptional(bool(o))
}

func (o AsCollectorOption) ApplyValueConsumer(b ValueConsumerBuilder) {
	b.SetAsCollector(bool(o))
}

func (o ByNameOption) ApplyValueConsumer(b ValueConsumerBuilder) {
	b.SetName(string(o))
}

func (o ByTagsOption) ApplyValueConsumer(b ValueConsumerBuilder) {
	b.AddTags(o.tags...)
}

func (o ScopeOption) ApplyValueConsumer(b ValueConsumerBuilder) {
	b.SetScope(o.scope)
}

func (o LocationOption) ApplyValueConsumer(b ValueConsumerBuilder) {
	b.SetLocation(o.Location)
}

func (o UpdateCallLocationOption) ApplyValueConsumer(b ValueConsumerBuilder) {
	b.UpdateCallLocation(o.Location)
}
