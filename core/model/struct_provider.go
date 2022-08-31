package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/reflecting"

	"github.com/jison/uni/internal/location"

	"github.com/jison/uni/internal/errors"
)

type structProvider struct {
	*structConsumer

	baseProvider
	com *component
}

var _ Provider = &structProvider{}

func (sp *structProvider) Components() ComponentCollection {
	return ComponentsOfIterator(sp.com)
}

func (sp *structProvider) Validate() error {
	errs := errors.Empty()

	if !reflecting.IsKindOrPtrOfKind(sp.sType, reflect.Struct) {
		return errors.Newf("[%v] is not type of struct or type of pointer of struct", sp.sType)
	}

	if err := sp.structConsumer.Validate(); err != nil {
		var structErr errors.StructError
		if errors.As(err, &structErr) {
			errs = structErr
		} else {
			errs = errs.AddErrors(err)
		}
	}

	if err := sp.com.Validate(); err != nil {
		errs = errs.AddErrors(err)
	}

	if errs.HasError() {
		return errs.WithMainf("%+v", sp)
	}

	return nil
}

func (sp *structProvider) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "Struct[%v] in %v", sp.sType, sp.Scope())

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " at %v", sp.Location())
	}
}

func (sp *structProvider) clone() *structProvider {
	cloned := &structProvider{
		baseProvider:   sp.baseProvider,
		structConsumer: sp.structConsumer.clone(),
		com:            sp.com.clone(),
	}
	for _, f := range cloned.structConsumer.fields {
		f.consumer = cloned
	}
	for _, f := range cloned.structConsumer.fakeFields {
		f.consumer = cloned
	}
	cloned.com.provider = cloned

	return cloned
}

func (sp *structProvider) Equal(other interface{}) bool {
	o, ok := other.(*structProvider)
	if !ok {
		return false
	}

	if sp.structConsumer != nil {
		if !sp.structConsumer.Equal(o.structConsumer) {
			return false
		}
	} else if o.structConsumer != nil {
		return false
	}

	if sp.com != nil {
		if !sp.com.Equal(o.com) {
			return false
		}
	} else if o.com != nil {
		return false
	}

	return true
}

type StructProviderBuilder interface {
	ModuleOption
	ProviderBuilder
	Field(fieldName string, opts ...DependencyOption) StructProviderBuilder
	IgnoreFields(predicate func(field reflect.StructField) bool) StructProviderBuilder

	SetIgnore(ignore bool) StructProviderBuilder
	SetHidden(hidden bool) StructProviderBuilder
	AddAs(ifs ...TypeVal) StructProviderBuilder
	SetName(name string) StructProviderBuilder
	AddTags(tags ...Symbol) StructProviderBuilder

	SetScope(scope Scope) StructProviderBuilder
	SetLocation(loc location.Location) StructProviderBuilder
	UpdateCallLocation(loc location.Location) StructProviderBuilder
}

func (sp *structProvider) ApplyModule(mb ModuleBuilder) {
	mb.AddProvider(sp)
}

func (sp *structProvider) Provider() Provider {
	return sp.clone()
}

func (sp *structProvider) Field(fieldName string, opts ...DependencyOption) StructProviderBuilder {
	sp.structConsumer.Field(fieldName, opts...)
	return sp
}

func (sp *structProvider) IgnoreFields(predicate func(field reflect.StructField) bool) StructProviderBuilder {
	sp.structConsumer.IgnoreFields(predicate)
	return sp
}

func (sp *structProvider) SetIgnore(ignore bool) StructProviderBuilder {
	sp.com.SetIgnore(ignore)
	return sp
}

func (sp *structProvider) SetHidden(hidden bool) StructProviderBuilder {
	sp.com.SetHidden(hidden)
	return sp
}

func (sp *structProvider) AddAs(ifs ...TypeVal) StructProviderBuilder {
	sp.com.AddAs(ifs...)
	return sp
}

func (sp *structProvider) SetName(name string) StructProviderBuilder {
	sp.com.SetName(name)
	return sp
}

func (sp *structProvider) AddTags(tags ...Symbol) StructProviderBuilder {
	sp.com.AddTags(tags...)
	return sp
}

func (sp *structProvider) SetScope(scope Scope) StructProviderBuilder {
	sp.baseConsumer.SetScope(scope)
	return sp
}

func (sp *structProvider) SetLocation(loc location.Location) StructProviderBuilder {
	sp.structConsumer.SetLocation(loc)
	return sp
}

func (sp *structProvider) UpdateCallLocation(loc location.Location) StructProviderBuilder {
	if sp.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		sp.SetLocation(loc)
	}
	return sp
}

func structProviderOf(t reflect.Type, opts ...StructProviderOption) *structProvider {
	sp := &structProvider{
		structConsumer: structConsumerOf(t),
		baseProvider:   baseProvider{},
		com: &component{
			val:   valuer.Identity(),
			rType: t,
		},
	}
	sp.com.provider = sp

	for _, f := range sp.fields {
		f.consumer = sp
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyStructProvider(sp)
	}

	return sp
}

func Struct(t TypeVal, opts ...StructProviderOption) StructProviderBuilder {
	sp := structProviderOf(TypeOf(t), opts...).UpdateCallLocation(nil)
	return sp
}

type StructProviderOption interface {
	ApplyStructProvider(StructProviderBuilder)
}

func (o FieldOption) ApplyStructProvider(b StructProviderBuilder) {
	b.Field(o.name, o.opts...)
}

func (o IgnoreFieldsOption) ApplyStructProvider(b StructProviderBuilder) {
	b.IgnoreFields(o.predicate)
}

func (o LocationOption) ApplyStructProvider(b StructProviderBuilder) {
	b.SetLocation(o.Location)
}

func (o UpdateCallLocationOption) ApplyStructProvider(b StructProviderBuilder) {
	b.UpdateCallLocation(o.Location)
}

func (o ScopeOption) ApplyStructProvider(b StructProviderBuilder) {
	b.SetScope(o.scope)
}

func (o IgnoreOption) ApplyStructProvider(b StructProviderBuilder) {
	b.SetIgnore(o.ignore)
}

func (o HiddenOption) ApplyStructProvider(b StructProviderBuilder) {
	b.SetHidden(o.hidden)
}

func (o AsOption) ApplyStructProvider(b StructProviderBuilder) {
	b.AddAs(o.ifs...)
}

func (o NameOption) ApplyStructProvider(b StructProviderBuilder) {
	b.SetName(string(o))
}

func (o TagsOption) ApplyStructProvider(b StructProviderBuilder) {
	b.AddTags(o.tags...)
}
